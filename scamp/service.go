package scamp

import "errors"
import "net"
import "crypto/tls"
import "fmt"

type ServiceAction func(Request,*Session)

type Service struct {
	serviceSpec   string
	name          string

	listener      net.Listener

	actions       map[string]ServiceAction
	sessChan      (chan *Session)
	isRunning     bool
	openConns     []*Connection
}

func NewService(serviceSpec string, name string) (serv *Service, err error){
	serv = new(Service)
	serv.name = name
	serv.serviceSpec = serviceSpec

	serv.actions = make(map[string]ServiceAction)
	serv.sessChan = make(chan *Session, 100)

	err = serv.listen()
	if err != nil {
		return
	}

	return
}

func (serv *Service)listen() (err error) {
	crtPath := config.ServiceCertPath(serv.name)
	keyPath := config.ServiceKeyPath(serv.name)

	if crtPath == nil || keyPath == nil {
		err = fmt.Errorf( "could not find valid crt/key pair for service %s", serv.name )
		return
	}

	cert, err := tls.LoadX509KeyPair( string(crtPath), string(keyPath) )
	if err != nil {
		return
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{ cert },
	}

	Trace.Printf("starting service on %s", serv.serviceSpec)
	serv.listener,err = tls.Listen("tcp", serv.serviceSpec, config)
	if err != nil {
		return err
	}

	return
}

func (serv *Service)Register(name string, action ServiceAction) (err error) {
	if serv.isRunning {
		return errors.New("cannot register handlers while server is running")
	}

	serv.actions[name] = action

	return
}

func (serv *Service)Run() {
	go serv.RouteSessions()

	for {
		netConn,err := serv.listener.Accept()
		if err != nil {
			Info.Printf("accept returned error. exiting service Run()")
			break
		}

		var tlsConn (*tls.Conn) = (netConn).(*tls.Conn)
		if tlsConn == nil {
			Error.Fatalf("could not create tlsConn")
			break
		}

		conn,err := newConnection(tlsConn, serv.sessChan)
		if err != nil {
			Error.Fatalf("error with new connection: `%s`", err)
			break
		}
		serv.openConns = append(serv.openConns, conn)

		go conn.packetRouter(false, true)
	}

	close(serv.sessChan)
}

// Spawn a router for each new session received over sessChan
func (serv *Service)RouteSessions() (err error){

	for newSess := range serv.sessChan {
		
		// if !stillOpen {
		// 	Trace.Printf("sessChan was closed. server is probably shutting down.")
		// 	break
		// }

		go func(){
			var action ServiceAction

			Trace.Printf("waiting for request to be received")
			request,err := newSess.RecvRequest()
			if err != nil {
				return
			}
			Trace.Printf("request came in for action `%s`", request.Action)

			action = serv.actions[request.Action]
			if action != nil {
				action(request, newSess)
				newSess.Free()
			} else {
				Error.Printf("unknown action `%s`", request.Action)
				// TODO: need to respond with 'unknown action'
			}
		}()
	}

	return
}

func (serv *Service)Stop(){
	// Sometimes we Stop() before service after service has been init but before it is started
	// The usual case is a bad config in another plugin
	if serv.listener != nil {
		serv.listener.Close()
	}
	for _,conn := range serv.openConns {
		conn.Close()
	}
}
