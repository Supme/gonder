package smtpSender

import (
	"errors"
	"sync"
)

// Config profile for sender pool
type Config struct {
	Hostname   string
	Iface      string
	Port       int
	Stream     int
	MapIP      map[string]string
	SMTPserver *SMTPserver
}

// Pipe email pipe for send email
type Pipe struct {
	wg     sync.WaitGroup
	email  chan Email
	config []Config
}

var ErrPipeStopped = errors.New("email streaming pipe stopped")

// NewPipe return new stream sender pipe
func NewPipe(conf ...Config) Pipe {
	pipe := Pipe{}
	pipe.config = append(pipe.config, conf...)
	return pipe
}

// Start stream sender
func (pipe *Pipe) Start() {
	pipe.wg = sync.WaitGroup{}
	pipe.email = make(chan Email, len(pipe.config))

	go func() {
		for i := range pipe.config {

			pipe.wg.Add(1)
			go func(conf *Config) {
				backet := make(chan struct{}, conf.Stream)
				for email := range pipe.email {
					backet <- struct{}{}
					pipe.wg.Add(1)
					go func(e Email) {
						conn := new(Connect)
						conn.SetHostName(conf.Hostname)
						conn.SetSMTPport(conf.Port)
						conn.SetIface(conf.Iface)
						conn.mapIP = conf.MapIP
						e.Send(conn, conf.SMTPserver)
						<-backet
						pipe.wg.Done()
					}(email)
				}
				pipe.wg.Done()
			}(&pipe.config[i])

		}
	}()
}

// Send add email to stream
func (pipe *Pipe) Send(email Email) (err error) {
	defer func(eml *Email, err *error) {
		if e := recover(); e != nil {
			*err = ErrPipeStopped
			//eml.ResultFunc(Result{ID: eml.ID, Err: errors.New("421 email streaming pipe stopped")})
		} else {
			*err = nil
		}
	}(&email, &err)
	pipe.email <- email
	return
}

// Stop stream sender
func (pipe *Pipe) Stop() {
	close(pipe.email)
	pipe.wg.Wait()
}

// NewEmailPipe return new chanel for stream send
// Deprecated: use NewPipe
func NewEmailPipe(conf ...Config) chan<- Email {
	pipe := NewPipe(conf...)
	pipe.Start()
	return pipe.email
}
