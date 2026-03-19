package seeding

//
//import (
//	"context"
//	"fmt"
//	"log"
//	"uniscore-seeding-bot/internal/domain/model"
//)
//
//// Service orchestrates seeding logic.
//type Service struct {
//	message       MessageGenerator
//	qualityFilter QualityFilter
//	persona       PersonaSelector
//	//publisher
//	// model Context bundle?
//}
//
//func NewService(mg MessageGenerator, qualityFilter QualityFilter, persona PersonaSelector) *Service {
//	return &Service{
//		message:       mg,
//		qualityFilter: qualityFilter,
//		persona:       persona,
//	}
//}
//
//func (s *Service) ProcessEvent(event model.MatchEvent, bundle model.ContextBundle) {
//	//1 chay template
//	draftMsg, _ := s.message.GenerateMessage(bundle, s.persona.SelectPersona(model.Match{ID: hashMessage()}))?
//
//	// chayfilter
//	ctx:=context.Background()
//	qResults,err:=s.qualityFilter.Check(event,ctx,*draftMsg,bundle)
//	if err!=nil{
//		retur nil,fmt.Errorf("ko lauy duoc ")
//	}
//
//	if !qResults.IsPass {
//		log.Warnf("Message bị chặn. Reason: %s, Action: %s", qResult.Reason, qResult.Action)
//		switch qResults.Action {
//		case "skip":
//			return
//		}
//		case "retry_p1":
//			draftMsg,_:=s.message.GenerateMessage(bundle, s.persona.SelectPersona(model.Match{ID: hashMessage()}))
//			case "retry_p2":
//				// LLm sinh lai
//				draftMsg,_:=s.message.GenerateMessageLLM(bundle, s.persona.SelectPersona())
//	}
//	// Check lại lần 2 (có giới hạn số lần retry)
//	qResult = s.qualityFilter.Check(event, draftMsg)
//	if !qResult.Pass { return } // Fail nữa thì bỏ qua luôn
//
//	//neu pass thi dua vao publisher 3-15s
//	//s.publish.PublisherWithDelay(draftMsg,3,15)
//}
