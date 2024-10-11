package spinner

import (
	"context"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestNewSpinner(t *testing.T) {
	s := New()
	if s.title != "Loading..." {
		t.Errorf("Expected default title 'Loading...', got '%s'", s.title)
	}
	if !reflect.DeepEqual(s.spinner.Spinner, spinner.Dot) {
		t.Errorf("Expected default spinner type to be Dot, got %v", s.spinner.Spinner)
	}
}

func TestSpinnerType(t *testing.T) {
	s := New().Type(Dots)
	if !reflect.DeepEqual(s.spinner.Spinner, spinner.Dot) {
		t.Errorf("Expected spinner type to be Dot, got %v", s.spinner.Spinner)
	}
}

func TestSpinnerDifferentTypes(t *testing.T) {
	s := New().Type(Line)
	if !reflect.DeepEqual(s.spinner.Spinner, spinner.Line) {
		t.Errorf("Expected spinner type to be Line, got %v", s.spinner.Spinner)
	}
}

func TestSpinnerView(t *testing.T) {
	s := New().Title("Test")
	view := s.View()

	if !strings.Contains(view, "Test") {
		t.Errorf("Expected view to contain title 'Test', got '%s'", view)
	}
}

func TestSpinnerContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := New().Context(ctx)
	cancel() // Cancel before running

	err := s.Run()
	if err != nil {
		t.Errorf("Run() returned an error after context cancellation: %v", err)
	}
}

func TestSpinnerStyleMethods(t *testing.T) {
	s := New()
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))

	s.Style(style)
	s.TitleStyle(titleStyle)

	if !reflect.DeepEqual(s.spinner.Style, style) {
		t.Errorf("Style was not set correctly")
	}

	if !reflect.DeepEqual(s.titleStyle, titleStyle) {
		t.Errorf("TitleStyle was not set correctly")
	}
}

func TestSpinnerInit(t *testing.T) {
	s := New()
	cmd := s.Init()

	if cmd == nil {
		t.Errorf("Init did not return a valid command")
	}
}

func TestSpinnerUpdate(t *testing.T) {
	s := New()
	cmd := s.Init()
	if cmd == nil {
		t.Errorf("Init did not return a valid command")
	}

	model, cmd := s.Update(spinner.TickMsg{})
	if reflect.TypeOf(model) != reflect.TypeOf(&Spinner{}) {
		t.Errorf("Update did not return correct model type")
	}

	if cmd == nil {
		t.Errorf("Update should return a non-nil command in this scenario")
	}

	// Simulate key press
	_, cmd = s.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Errorf("Update did not handle key press correctly")
	}
}

func TestAccessibleSpinner(t *testing.T) {
	s := New().Accessible(true)
	err := s.Run()
	if err != nil {
		t.Errorf("Run() in accessible mode returned an error: %v", err)
	}
}

func TestSpinnerWithoutSignalHandling(t *testing.T) {
	s := New().ProgramOptions(tea.WithoutSignalHandler())

	// Action function that simulates a long-running task
	a := func() {
		time.Sleep(10 * time.Second)
		t.Errorf("This should not be executed")
	}

	// Channel to capture OS signals
	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		// Run the spinner with the action
		done <- s.Action(a).Run()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// Wait for either the spinner to complete or a signal
	select {
	case <-sigs:
		// If we received a signal, pass the test and stop the spinner
		t.Log("Received SIGINT as expected")
	case err := <-done:
		// If the spinner completed without being interrupted, fail the test
		if err == nil {
			t.Errorf("Expected Run() to be interrupted by SIGINT, but it completed without error")
		} else {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
