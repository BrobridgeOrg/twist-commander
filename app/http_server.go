package app

import (
	"context"
	"net/http"

	pb "twist-commander/pb"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

type TaskAction struct {
	Type    string            `json:"type"`
	Method  string            `json:"method"`
	Uri     string            `json:"uri"`
	Headers map[string]string `json:"headers"`
	Payload string            `json:"payload"`
}

type Task struct {
	Actions map[string]TaskAction `json:"actions"`
}

type ConfirmTransactionRequest struct {
	TransactionID string `json:"transactionID"`
	Tasks         []Task `json:"tasks"`
	Expires       uint64 `json:"expires"`
}

func (a *App) InitHTTPServer(host string) error {

	lis := a.connectionListener.Match(cmux.HTTP1Fast())

	//	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Router
	r.POST("/api/transactions", func(c *gin.Context) {

		reply, err := a.grpcServer.Commander.CreateTransaction(context.Background(), &pb.CreateTransactionRequest{})
		if err != nil {

			c.JSON(400, gin.H{
				"success": false,
			})

			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"transactionID": reply.TransactionID,
		})
	})

	r.POST("/api/transactions/:transactionID", func(c *gin.Context) {

		var request ConfirmTransactionRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		in := &pb.ConfirmTransactionRequest{
			TransactionID: c.Param("transactionID"),
			//			Expires: request.Expires,
		}

		tasks := make([]*pb.TransactionTask, 0)
		for _, task := range request.Tasks {

			t := &pb.TransactionTask{}
			tasks = append(tasks, t)

			for name, action := range task.Actions {
				log.Info(action)
				act := &pb.TransactionTaskAction{
					Type:    action.Type,
					Method:  action.Method,
					Uri:     action.Uri,
					Headers: action.Headers,
					Payload: action.Payload,
				}
				if name == "confirm" {
					t.Confirm = act
				} else if name == "cancel" {
					t.Cancel = act
				}
			}
		}

		in.Tasks = tasks

		reply, err := a.grpcServer.Commander.ConfirmTransaction(context.Background(), in)
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       reply.Success,
			"transactionID": reply.TransactionID,
		})
	})

	s := &http.Server{
		Handler: r,
	}

	log.WithFields(log.Fields{
		"host": host,
	}).Info("Starting HTTP server on " + host)

	// Starting server
	if err := s.Serve(lis); err != cmux.ErrListenerClosed {
		log.Fatal(err)
		return err
	}

	return nil
}
