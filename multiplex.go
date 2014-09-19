package tbeer

import (
	"sync"
)

type Multiplexable func(out chan<- interface{}) error

// Generically multiplex several "producer" functions
// with proper error handling, meaning that the function
// blocks until all producers have successfully started
// producing, or at least one has failed to initialize.
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

// "Uniplex" - singular version of the multiplexer, which is simpler, uses
// less resources and has an identical type of error system.
//
// I have come to the conclusion that using a single channel for this
// is theoretically impossible:
// It is possible to read the first item and then put it back, but that
// would ruin the order of items.
//
// The reason for this problem is the complexity of capturing the event
// that a producer function does*not*return before a specific point, an event
// which in itself is a non-event - so we need to use the point instead.
func Uniplex(bufsize int, fn Multiplexable) (<-chan interface{}, error) {
	inner := make(chan interface{}, bufsize)
	outer := make(chan interface{}, bufsize)
	ec := make(chan error, 0)

	go func() {
		if err := fn(inner); err != nil {
			ec <- err
		}
		close(inner)
	}()

	select {
	case err := <-ec:
		return nil, err
	case item, ok := <-inner:
		if ok {
			// just forward all items to outer channel
			go func() {
				outer <- item
				for item := range inner {
					outer <- item
				}
				close(outer)
			}()

		} else {
			close(outer)
		}

		return outer, nil
	}
}
