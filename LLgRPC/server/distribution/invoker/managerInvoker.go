package invoker

import (
	"fmt"

	"github.com/lucas625/Middleware/LLgRPC/common/distribution/marshaller"
	"github.com/lucas625/Middleware/LLgRPC/common/distribution/packet"
	"github.com/lucas625/Middleware/LLgRPC/server/distribution/lifecycle-management/pooling"
	"github.com/lucas625/Middleware/LLgRPC/server/infrastructure/srh"
	"github.com/lucas625/Middleware/LLgRPC/server/service/database"
	"github.com/lucas625/Middleware/LLgRPC/server/service/manager"
	"github.com/lucas625/Middleware/LLgRPC/server/service/person"
)

// ManagerInvoker is a structure to run the invoker.
//
// Members:
//  none
//
type ManagerInvoker struct{}

func writeFromPool(pool *pooling.Pool) {
	man := pool.GetFromPool().(*manager.Manager)
	man.Write("files/")
}

// Invoke is a funcion to set the server running.
//
// Parameters:
//  none
//
// Returns:
//  none
//
func (ManagerInvoker) Invoke() {
	marshallerImpl := marshaller.Marshaller{}
	packetPacketReply := packet.Packet{}
	replParams := make([]interface{}, 1)

	// creating the pool
	db := database.InitDatabase()
	managerList := make([]interface{}, 11)
	for i := 0; i < len(managerList); i++ {
		managerAux := manager.Manager{DB: db}
		managerList[i] = &managerAux
	}
	manPool := pooling.InitPool(managerList)
	defer writeFromPool(manPool)
	defer pooling.EndPool(manPool)
	man := manPool.GetFromPool().(manager.Manager)
	man.Load("files/database.json")

	fmt.Println("Server invoking.")

	for {
		srhImpl := srh.SRH{ServerHost: "localhost", ServerPort: 8080}

		// Receive data
		rcvMsgBytes := (&srhImpl).Receive()

		// 	unmarshall
		packetPacketRequest := marshallerImpl.Unmarshall(rcvMsgBytes)

		// setup request
		var manA *manager.Manager
		manA = manPool.GetFromPool().(*manager.Manager)

		// finding the operation
		operation := packetPacketRequest.Bd.ReqHeader.Operation
		switch operation {
		case "Add":
			_p1 := packetPacketRequest.Bd.ReqBody.Body[0].(person.Person)
			manA.AddPerson(_p1)
			replParams[0] = nil
		case "Remove":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			replParams[0] = manA.RemovePerson(_p1)
		case "GetPerson":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			replParams[0] = manA.GetPerson(_p1)
		case "SetPerson":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			_p2 := packetPacketRequest.Bd.ReqBody.Body[1].(person.Person)
			replParams[0] = manA.SetPerson(_p1, _p2)
		case "GetName":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			replParams[0] = manA.GetName(_p1)
		case "GetAge":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			replParams[0] = manA.GetAge(_p1)
		case "GetGender":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			replParams[0] = manA.GetGender(_p1)
		case "SetName":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			_p2 := packetPacketRequest.Bd.ReqBody.Body[1].(string)
			replParams[0] = manA.SetName(_p1, _p2)
		case "SetAge":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			_p2 := int(packetPacketRequest.Bd.ReqBody.Body[1].(float64))
			replParams[0] = manA.SetAge(_p1, _p2)
		case "SetGender":
			_p1 := int(packetPacketRequest.Bd.ReqBody.Body[0].(float64))
			_p2 := packetPacketRequest.Bd.ReqBody.Body[1].(string)
			replParams[0] = manA.SetGender(_p1, _p2)
		}

		// assembly packet
		repHeader := packet.ReplyHeader{Context: "", RequestID: packetPacketRequest.Bd.ReqHeader.RequestID, Status: 1}
		repBody := packet.ReplyBody{OperationResult: replParams}
		header := packet.Header{Magic: "packet", Version: "1.0", ByteOrder: true, MessageType: 0} // MessageType 0 = reply
		body := packet.Body{RepHeader: repHeader, RepBody: repBody}
		packetPacketReply = packet.Packet{Hdr: header, Bd: body}

		// marshall reply
		msgToClientBytes := marshallerImpl.Marshall(packetPacketReply)

		// send Reply
		(&srhImpl).Send(msgToClientBytes)
	}
}
