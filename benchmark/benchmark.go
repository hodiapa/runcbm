package benchmark

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/washraf/runcbm/common"
	"github.com/washraf/runcbm/containers/config"
	"github.com/washraf/runcbm/containers/conutil"
)

//Run ..
func Run(containerID string, n int) error {
	condir, err := config.FindContainerBundle(containerID)
	if err != nil {
		return err
	}

	measuresList := make(Measures, 0)
	for i := 1; i <= n; i++ {
		delcommand := exec.Command("rm")
		delcommand.Dir = condir
		delcommand.Args = append(delcommand.Args, "-rf")
		delcommand.Args = append(delcommand.Args, "checkpoint/")
		r, err := delcommand.CombinedOutput()
		if err != nil {
			fmt.Println("Delete Checkpoint error")
			return err
		}
		fmt.Println("trial number ", i)
		//time.Sleep(time.Second * 5)
		fmt.Println("Sleep for 5 Seconds")
		time.Sleep(time.Second * 5)
		measure := Measure{}
		u, err := conutil.GetContainerUtilization(containerID)
		if err != nil {
			fmt.Println("Read Utilization error")
			return err
		}
		measure.ID = i
		measure.ProcessCount = u.ProcessCount
		measure.TaskCount = u.TaskCount
		measure.MemorySize = u.UsedMemory
		command := exec.Command("time", "-f", "%e", "runc", "checkpoint", "--tcp-established", "--empty-ns", "network", containerID)
		command.Dir = condir
		r, err = command.CombinedOutput()
		if err != nil {
			fmt.Println("Checkpoint error")
			return err
		}
		measure.CheckpointTime, _ = strconv.ParseFloat(strings.TrimSpace(string(r)), 64)
		s, err := common.FindDiskSizeMB(condir + "/checkpoint/")
		if err != nil {
			fmt.Println("Read Checkpoint Disk size error")

			fmt.Println(condir)
			return err
		}
		measure.Checkpointsize = s
		command = exec.Command("time", "-f", "%e", "runc", "restore", "-d", "--tcp-established", containerID)
		//command.Dir = "/containers/"+container+"/"
		command.Dir = condir

		r, err = command.CombinedOutput()
		if err != nil {
			fmt.Println("Restore error")

			return err
		}
		measure.Restoretime, _ = strconv.ParseFloat(strings.TrimSpace(string(r)), 64)
		measuresList = append(measuresList, measure)
		err = writetoFile(logFile, measure)
		if err != nil {
			return err
		}
	}
	printlist(measuresList)
	return nil
}

func printlist(measuresList Measures) {
	fmt.Printf("ID\tProcessCount\tTaskCount\tMemorySize\tCheckpointTime\tCheckpointsize\tRestoretime\n")
	for _, m := range measuresList {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%v\n", m.ID, m.ProcessCount, m.TaskCount, m.MemorySize, m.CheckpointTime, m.Checkpointsize, m.Restoretime)
	}
}
func writetoFile(filename string, m Measure) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.WriteString(strconv.FormatInt(int64(m.ID), 10))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatInt(int64(m.ProcessCount), 10))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatInt(int64(m.TaskCount), 10))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatInt(int64(m.MemorySize), 10))
	buffer.WriteString(",")
	buffer.WriteString(floatToString(m.CheckpointTime))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatInt(int64(m.Checkpointsize), 10))
	buffer.WriteString(",")
	buffer.WriteString(floatToString(m.Restoretime))
	buffer.WriteString("\n")

	_, err = f.WriteString(string(buffer.Bytes()))
	if err != nil {
		return err
	}
	f.Close()
	/*
		err := ioutil.WriteFile(filename, buffer.Bytes(), 0644)
	*/

	return nil
}

func floatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}
