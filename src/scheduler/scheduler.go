package scheduler

import (
	"log"
	"sync"
	"time"
)

// Task represents a scheduled task
type Task struct {
	Name     string
	Schedule string // Cron-like: "0 3 * * 0" = Sunday 3:00 AM
	Fn       func() error
	nextRun  time.Time
	running  bool
}

// Scheduler manages periodic tasks
type Scheduler struct {
	tasks []*Task
	stop  chan struct{}
	wg    sync.WaitGroup
	mu    sync.Mutex
}

// New creates a new scheduler
func New() *Scheduler {
	return &Scheduler{
		tasks: make([]*Task, 0),
		stop:  make(chan struct{}),
	}
}

// AddTask adds a new scheduled task
func (s *Scheduler) AddTask(name, schedule string, fn func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		Name:     name,
		Schedule: schedule,
		Fn:       fn,
		nextRun:  calculateNextRun(schedule),
	}

	s.tasks = append(s.tasks, task)
	log.Printf("Scheduler: Added task '%s' with schedule '%s'", name, schedule)
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Println("Scheduler: Started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
	log.Println("Scheduler: Stopped")
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.checkTasks(now)
		}
	}
}

// checkTasks checks if any tasks need to run
func (s *Scheduler) checkTasks(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if !task.running && now.After(task.nextRun) {
			go s.runTask(task)
		}
	}
}

// runTask executes a task
func (s *Scheduler) runTask(task *Task) {
	s.mu.Lock()
	task.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		task.running = false
		task.nextRun = calculateNextRun(task.Schedule)
		s.mu.Unlock()
	}()

	log.Printf("Scheduler: Running task '%s'", task.Name)

	if err := task.Fn(); err != nil {
		log.Printf("Scheduler: Task '%s' failed: %v", task.Name, err)
	} else {
		log.Printf("Scheduler: Task '%s' completed successfully", task.Name)
	}
}

// calculateNextRun calculates the next run time based on cron schedule
// Simplified version - supports weekly schedules like "0 3 * * 0"
func calculateNextRun(schedule string) time.Time {
	now := time.Now()

	// Parse schedule (simplified for weekly: "0 3 * * 0" = Sunday 3 AM)
	// For now, if contains "* * 0", it's weekly on Sunday
	if schedule == "0 3 * * 0" {
		// Next Sunday at 3:00 AM
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		if daysUntilSunday == 0 && now.Hour() >= 3 {
			daysUntilSunday = 7
		}

		nextSunday := now.AddDate(0, 0, daysUntilSunday)
		return time.Date(nextSunday.Year(), nextSunday.Month(), nextSunday.Day(),
			3, 0, 0, 0, now.Location())
	}

	// Default: run in 7 days
	return now.Add(7 * 24 * time.Hour)
}
