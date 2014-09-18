package tbeer

import (
	"sync"
)

type Multiplexable func(out chan<- interface{}) error

func Multiplex(bufsize int, fns ...Multiplexable) (<-chan interface{}, error) {
	out := make(chan interface{}, bufsize)
	cancel := make(chan struct{}, 0)
	setupErr := make(chan error, 0)
	wg := sync.WaitGroup{}
	wg.Add(len(fns))

	// wait for all producers to finish
	go func() {
		wg.Wait()
		close(out)
	}()

	for _, fn := range fns {
		// fire off the producers and multiplexers

		// first make a channel to communicate errors locally
		pErr := make(chan error, 0)
		// channel to transfer items from producer to multiplexer
		items := make(chan interface{}, 0)

		// run the producer in the following goroutine
		go func(fn Multiplexable) {
			if err := fn(items); err != nil {
				pErr <- err
			}
			close(items)
		}(fn)

		// run the multiplexing algorithm and state machine here
		go func() {
			// Syncrhonize closing when done
			defer wg.Done()

			// What happens first?
			select {
			case <-cancel:
				// we got cancelled before anything interesting happened
			case err := <-pErr:
				// an error occured before an item was produced, we interpret
				// this as a failure to set up the producer routine
				setupErr <- err
			case item, ok := <-items:
				// item was produced, thus setup succeeded
				setupErr <- nil

				if ok {
					// multiplex first item
					out <- item

					for {
						select {
						case <-cancel:
							return
						case <-pErr:
							// non-fatal error..
							// should we stop this producer?
							return
						case item, ok := <-items:
							if ok {
								// multiplex item
								out <- item
							} else {
								// closed
								return
							}
						}
					}
				}
			}
		}()
	}

	for i := 0; i < len(fns); i++ {
		// collect setup status from all producers
		if err := <-setupErr; err != nil {
			close(cancel)
			return nil, err
		}
	}

	return out, nil
}
