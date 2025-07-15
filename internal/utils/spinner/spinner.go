package spinner

import (
	"time"

	"github.com/briandowns/spinner"
)

// StartSpinner starts a terminal spinner with the given message.
// Returns a stop function to halt and clear the spinner.
//
// Usage: assign the spinner to a 'stop' variable, run some code, then call stop().
// i.e.:
//
//	stop := spinner.StartSpinner("Your message here ")
//	err := lib.SomeOperation()
//	stop()
//	if err != nil { return err }
func StartSpinner(message string) func() {
	// Use a nice character set; CharSets[14] is a good default.
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()

	// Return a function to stop the spinner and print "Done!"
	return func() {
		s.Stop()
		// Optionally, print a newline or a "Done!" message.
		// fmt.Println("Done!")
	}
}
