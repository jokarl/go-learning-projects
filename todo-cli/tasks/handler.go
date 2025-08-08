package tasks

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	path string // Path to the task file
}

type TaskHandler interface {
	Add(task string) error
	Delete(ids []int) ([]int, error)
	Complete(ids []int) ([]int, error)
	GetAll() []Task
}

type HandlerOptions struct {
	Path string // Path to the task file
}

// NewHandler creates a new Handler instance for managing tasks.
func NewHandler(opts *HandlerOptions) (*Handler, error) {
	h := &Handler{
		path: opts.Path,
	}

	err := h.createFile()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Handler) deleteFile() error {
	if _, err := os.Stat(h.path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %s does not exist", h.path)
	}

	return os.Remove(h.path)
}

func (h *Handler) createFile() error {
	t := reflect.TypeOf(Task{})
	if _, err := os.Stat(h.path); errors.Is(err, os.ErrNotExist) {
		f, err2 := os.OpenFile(h.path, os.O_CREATE|os.O_WRONLY, 0644)
		if err2 != nil {
			return fmt.Errorf("error creating file: %v", err2)
		}

		var columns []string
		for i := 0; i < t.NumField(); i++ {
			columns = append(columns, t.Field(i).Name)
		}
		_, err := f.WriteString(strings.Join(columns, ",") + "\n")
		if err != nil {
			return err
		}
	}
	// File already exists, verify number of columns
	r, err := h.readCsv()
	if err != nil {
		return err
	}

	if len(r[0]) == t.NumField() {
		return nil // File already has the correct number of columns
	} else {
		_ = fmt.Sprintf("Error: file %s has unexpected number of coulmns (expected %d)", h.path, t.NumField())
		return err
	}
}

// Delete deletes a task.
// Returns a slice of IDs that were successfully deleted.
// If no IDs are provided, it returns an error.
func (h *Handler) Delete(ids []int) ([]int, error) {
	// Sanity check that IDs are provided
	if len(ids) == 0 {
		return nil, fmt.Errorf("received no IDs to delete")
	}

	fin, err := os.Open(h.path)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(fin)
	var lineNum = 0
	var deleteIds []int
	for scanner.Scan() {
		line := scanner.Text()
		if lineNum == 0 {
			lines = append(lines, line) // Always keep header
		} else {
			i, err := strconv.Atoi(strings.Split(line, ",")[0])
			if err != nil {
				lineNum++
				continue // Skip lines with non-integer ID
			}
			if slices.Contains(ids, i) {
				deleteIds = append(deleteIds, i) // Track deleted ID
				lineNum++
				continue
			}
			lines = append(lines, line) // Keep line
		}
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Overwrite existing file with lines to keep
	fout, err := os.OpenFile(h.path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	for _, l := range lines {
		_, err := fout.WriteString(l + "\n")
		if err != nil {
			return nil, err
		}
	}

	if err := fout.Close(); err != nil {
		return nil, err
	}

	return deleteIds, nil
}

type CompletionResult struct {
	CompleteIds         []int // IDs of tasks that were successfully completed
	AlreadyCompletedIds []int // IDs of tasks that were already completed
	NotFoundIds         []int // IDs of tasks that were not found
}

// Complete marks tasks as completed.
func (h *Handler) Complete(ids []int) (*CompletionResult, error) {
	// Sanity check that IDs are provided
	if len(ids) == 0 {
		return nil, fmt.Errorf("received no IDs to complete")
	}

	fin, err := os.Open(h.path)
	if err != nil {
		return nil, err
	}

	var result = &CompletionResult{
		CompleteIds:         []int{},
		AlreadyCompletedIds: []int{},
		NotFoundIds:         []int{},
	}

	var statusIndex = -1
	t := reflect.TypeOf(Task{})
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type == reflect.TypeOf(TaskState(0)) {
			statusIndex = i
			break
		}
	}

	if statusIndex == -1 {
		return nil, fmt.Errorf("could not find status field")
	}
	fmt.Println("Status index is ", statusIndex)

	var lines []string
	scanner := bufio.NewScanner(fin)
	var lineNum = 0
	for scanner.Scan() {
		line := scanner.Text()
		if lineNum == 0 {
			lines = append(lines, line) // Always keep header
		} else {
			columns := strings.Split(line, ",")
			i, err := strconv.Atoi(columns[0])
			if err != nil {
				lineNum++
				continue // Skip lines with non-integer ID
			}
			// The task should be marked as completed if its ID is in the provided list
			if slices.Contains(ids, i) {
				// If the task is already completed, track it separately
				if columns[statusIndex] == Completed.String() {
					result.AlreadyCompletedIds = append(result.AlreadyCompletedIds, i)
					lines = append(lines, line) // Keep line as is
					lineNum++
					continue
				}
				// Task is not completed, mark it as completed
				columns[statusIndex] = Completed.String()
				result.CompleteIds = append(result.CompleteIds, i) // Track completed ID
				lines = append(lines, strings.Join(columns, ","))  // Keep line
				lineNum++
				continue
			} else {
				lines = append(lines, line) // Keep line as is
			}
		}
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Overwrite existing file with lines to keep
	fout, err := os.OpenFile(h.path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	for _, l := range lines {
		_, err := fout.WriteString(l + "\n")
		if err != nil {
			return nil, err
		}
	}

	if err := fout.Close(); err != nil {
		return nil, err
	}

	result.NotFoundIds = slices.Concat(result.CompleteIds, result.AlreadyCompletedIds)
	return result, nil
}

// Add adds a task.
func (h *Handler) Add(d string) error {
	tasks := h.GetAll()
	id := 0
	for _, task := range tasks {
		if task.Id > id {
			id = task.Id
		}
	}

	return h.appendLine(&Task{
		Id:          id + 1,
		Description: d,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		Status:      Pending,
	})
}

// appendLine appends a new line to the task file.
// It takes an ID and a description as parameters.
// The ID is the next available ID, and the description is the task description.
func (h *Handler) appendLine(task *Task) error {
	f, err := os.OpenFile(h.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	line := fmt.Sprintf("%d,%s,%s,%s\n", task.Id, task.Description, task.CreatedAt, task.Status)
	if _, err := f.Write([]byte(line)); err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

// GetAll retrieves all headers and tasks from the task file.
func (h *Handler) GetAll() []Task {
	records, err := h.readCsv()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	t := h.mapRecords(records)
	return t
}

// readCsv reads the CSV file and returns all records.
func (h *Handler) readCsv() ([][]string, error) {
	file, err := os.Open(h.path)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}(file)

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

// mapRecords maps the CSV records to a slice of Task structs.
func (h *Handler) mapRecords(records [][]string) []Task {
	var t []Task
	for i, record := range records {
		// Ignore headers
		if i == 0 {
			continue
		}
		id, err := strconv.Atoi(record[0])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		var status TaskState
		for k, v := range stateName {
			if v == record[3] {
				status = k
				break
			}
		}

		t = append(t, Task{
			Id:          id,
			Description: record[1],
			CreatedAt:   record[2],
			Status:      status,
		})
	}
	return t
}
