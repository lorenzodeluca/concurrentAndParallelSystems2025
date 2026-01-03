// i'll try to comment the code in italian because in the exam i'll need to write it in italian        ＼（〇_ｏ）／
package main

import (
	"fmt"
	"math/rand"
	"time"
)

const NP = 3  // weights room max capacity
const MAX = 5 // gym max total capacity
const NT = 3  //  number of PT
const USERS = 7
const MAXBUFF = 100
const MAXPROC = USERS + NT
const NUM_AREE = 2
const PESIAREA = 0
const CORSIAREA = 1

var ingressoArea [NUM_AREE]chan Richiesta
var uscita = make(chan Richiesta, MAXBUFF)
var ingressoAreaPT chan Richiesta //priority access
var uscitaAreaPT chan Richiesta   //si potrebbe fare con un solo canale di uscita ma averne uno separato per i PT semplifica sensibilmente la logica

var closeCenterCmd = make(chan bool)
var endServerCmd = make(chan bool)
var done = make(chan bool)

type Richiesta struct {
	id   int
	tipo int
	ack  chan bool
}

type Trainer struct {
	dentro          bool
	vuoleUscire     bool
	utenteAssegnato int
	ackUscita       chan bool
}

func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

func user(id int) {
	fmt.Printf("[user %d] ready to enter the gym\n", id)
	for {
		//tipo = 0 means the user wants to exit the center
		//tipo = 1 means the user wants to enter the courses area
		//tipo = 2 means the user wants to enter the weights area
		var ric Richiesta
		ric.id = id
		ric.tipo = rand.Intn(NUM_AREE + 1) // range 0-2 inclusive (NUM_AREE==2)
		ric.ack = make(chan bool, MAXBUFF)

		fmt.Printf("[user %d] choose the area %d\n", id, ric.tipo)
		if ric.tipo == NUM_AREE {
			break
		}

		//area entrance procedure
		if ric.tipo == 1 {
			ingressoArea[CORSIAREA] <- ric
		}
		if ric.tipo == 0 {
			ingressoArea[PESIAREA] <- ric
		}
		<-ric.ack // waiting for server ack

		fmt.Printf("[user %d] entered the the area %d \n", id, ric.tipo)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

		//area exit procedure
		uscita <- ric
		<-ric.ack // waiting for server ack

		fmt.Printf("[user %d] left the the area %d \n", id, ric.tipo)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}
	fmt.Printf("[user %d] left the gym\n", id)
	done <- true
}

func trainer(id int) {
	fmt.Printf("[trainer %d] entered the gym center\n", id)

	var ric Richiesta
	ric.id = id
	ric.tipo = CORSIAREA // range 0-NUM_AREE inclusive
	ric.ack = make(chan bool, MAXBUFF)

	for {
		fmt.Printf("[trainer %d] entering the courses area \n", id)

		//area entrance procedure
		ingressoAreaPT <- ric

		if <-ric.ack == false { // waiting for server ack
			//center closing
			break
		}

		fmt.Printf("[trainer %d] entered the courses area \n", id)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

		//area exit procedure
		uscita <- ric
		<-ric.ack // waiting for server ack

		fmt.Printf("[trainer %d] left the the area \n", id)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}
	fmt.Printf("[trainer %d] left the spa center\n", id)
	done <- true
}

// ritorna l'id del pt libero oppure -1
func trainerLibero(trainers map[int]Trainer) int {
	for ptId, trainerInfo := range trainers {
		if trainerInfo.dentro && trainerInfo.utenteAssegnato == -1 {
			return ptId // id del pt libero
		}
	}
	return -1 // non c'è nessun pt libero
}

// goroutine used to control the access to the center and its areas
func server() {
	capCounter := 0 //users inside center
	centerClosing := false

	trainers := make(map[int]Trainer) //stato dei trainer, mappa con come chiave l'id del trainer

	//inizializzando i trainers a fuori
	for ptId, trainerInfo := range trainers {
		trainerInfo.dentro = false
		trainerInfo.utenteAssegnato = -1
		trainerInfo.vuoleUscire = false
		trainers[ptId] = trainerInfo
	}

	for {
		select {
		// personal trainers always have priority
		case ric := <-ingressoAreaPT:
			trainerToEdit := trainers[ric.id]
			trainerToEdit.dentro = true // nessun utente assegnato, aggiunto il trainer alla mappa per indicare che è disponibile a dare lezione
			trainerToEdit.vuoleUscire = false
			trainerToEdit.utenteAssegnato = -1
			trainers[ric.id] = trainerToEdit
			if centerClosing {
				ric.ack <- false
			} else {
				ric.ack <- true
			}

		//entrata utenti area corsi, solo se c'è un Personal Trainer libero e c'è posto dentro la palestra
		case ric := <-when(capCounter < MAX && len(ingressoAreaPT) == 0 && trainerLibero(trainers) != -1, ingressoArea[CORSIAREA]):
			capCounter++
			idTrainerLibero := trainerLibero(trainers)
			trainerLibero := trainers[idTrainerLibero]
			trainerLibero.utenteAssegnato = ric.id
			trainers[idTrainerLibero] = trainerLibero
			ric.ack <- true

		//entrata utenti area pesi, priorità a utenti corsi e PT
		case r := <-when(capCounter < MAX && len(ingressoArea[CORSIAREA]) == 0 && len(ingressoAreaPT) == 0, ingressoArea[PESIAREA]):
			capCounter++
			r.ack <- true

		//uscita utenti
		case r := <-uscita:
			if r.tipo == CORSIAREA {
				// libero il PT
				for ptId, trainerInfo := range trainers {
					if trainerInfo.utenteAssegnato == r.id {
						trainerInfo.utenteAssegnato = -1
						if trainerInfo.vuoleUscire {
							trainerInfo.ackUscita <- true
							trainerInfo.dentro = false
						}
						trainers[ptId] = trainerInfo
						break
					}
				}
			}
			capCounter--
			r.ack <- true

		//uscita pt se il pt non sta lavorando
		case r := <-uscitaAreaPT:
			if r.tipo == CORSIAREA {
				if trainers[r.id].utenteAssegnato == -1 {
					r.ack <- true // non sta lavorando posso autorizzarlo ad uscire
				} else {
					trainerToEdit := trainers[r.id]
					trainerToEdit.vuoleUscire = true
					trainerToEdit.ackUscita = r.ack
					trainers[r.id] = trainerToEdit
				}
			}

		case <-closeCenterCmd: // when all users ended
			centerClosing = true

		case <-endServerCmd: // when all routines(users+lifeguards) ended
			fmt.Println("THE END !!!!!!")
			done <- true
			return
		}
	}
}

func main() {
	//init canali
	for i := 0; i < len(ingressoArea); i++ {
		ingressoArea[i] = make(chan Richiesta, MAXBUFF)
	}
	ingressoAreaPT = make(chan Richiesta, MAXBUFF)
	uscitaAreaPT = make(chan Richiesta, MAXBUFF)

	go server()
	for i := 0; i < NT; i++ {
		go trainer(i)
	}
	for i := 0; i < USERS; i++ {
		go user(i)
	}

	for i := 0; i < USERS; i++ {
		<-done
	}
	closeCenterCmd <- true
	for i := 0; i < NT; i++ {
		<-done
	}
	endServerCmd <- true //telling server to shut down
	<-done               //waiting for server confirmation

	fmt.Printf("\n HO FINITO ")
}
