package main

import (
	"fmt"
	"math/rand"
	"time"
)

// limiti
const CLIENTI_MAX_PER_DIPENDENTI = 3
const MAX = 6
const NM = 3 // dimensione lotto mascherine

// info simulazione
const N_DIPENDENTI = 2
const N_CLIENTI = 10
const N_FORNITORI = 1

// altre variabili
const MAXBUFF = 100
const MAXPROC = N_DIPENDENTI + N_CLIENTI + N_FORNITORI //per sapere quanti proc in esecuzione oltre al server
const ID_CLIENTE = 0
const ID_CLIENTE_ABITUALE = 1
const ID_COMMESSO = 2
const ID_FORNITORE = 3
const N_TIPOLOGIE_UTENTI_CHE_ACCEDONO = 3 // FORNITORI NON ACCEDONO
const TEMPO_ATTESA_MAX = 5

// risorse/canali
var ingresso [N_TIPOLOGIE_UTENTI_CHE_ACCEDONO]chan Richiesta
var ottenimentoMascherina chan Richiesta
var consegnaMascherine chan Richiesta
var uscita chan Richiesta

var closeSignalCmd = make(chan bool)
var endServerCmd = make(chan bool)
var done = make(chan bool)

// classi/strutture utili
type Richiesta struct {
	id   int //id utente
	tipo int
	ack  chan int
}

func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

func idToString(tipo int) string {
	switch tipo {
	case ID_CLIENTE:
		return "CLIENTE"
	case ID_CLIENTE_ABITUALE:
		return "CLIENTE ABITUALE"
	case ID_COMMESSO:
		return "DIPENDENTE"
	default:
		return "errore"
	}
}

func cliente(id int, tipo int) {
	fmt.Printf("[user %d/%s] vuole entrare\n", id, idToString(tipo))

	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.ack = make(chan int, MAXBUFF)

	//utente prende mascherina
	ottenimentoMascherina <- ric
	<-ric.ack // waiting for server ack
	fmt.Printf("[user %d/%s] ha ottenuto la mascherina\n", id, idToString(tipo))

	time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second) //4. visita del castello in un tempo arbitrario

	//utente manda richiesta di ingresso
	ingresso[tipo] <- ric
	matricolaCommesso := <-ric.ack // waiting for server ack
	fmt.Printf("[user %d/%s] è entrato supervisionato dal commesso %d\n", id, idToString(tipo), matricolaCommesso)

	time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second) //4. visita del castello in un tempo arbitrario

	fmt.Printf("[user %d/%s] è pronto ad uscire \n", id, idToString(tipo))

	uscita <- ric
	<-ric.ack // waiting for server ack	fmt.Printf("[user %d/%s] 6. percorre la strada in discesa impiegando un tempo arbitrario \n", id, idToString(tipo))

	fmt.Printf("[user %d/%s] è uscito\n", id, idToString(tipo))
	done <- true
}

func commesso(id int) {
	tipo := ID_COMMESSO
	fmt.Printf("[commesso %d] si è svegliato\n", id)

	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.ack = make(chan int, MAXBUFF)

	richiestaUscita := 0
	for {
		fmt.Printf("[commesso %d] vuole entrare\n", id)
		ingresso[tipo] <- ric
		richiestaUscita = <-ric.ack // waiting for server ack

		if richiestaUscita == -1 {
			break
		}

		fmt.Printf("[commesso %d] sta sorvegliando\n", id)
		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)
		fmt.Printf("[commesso %d] non ha più voglia di lavorare e chiede al manager di uscire\n", id)

		uscita <- ric
		<-ric.ack // waiting for server ack
		fmt.Printf("[commesso %d] è stato autorizzato ad uscire ed è uscito\n", id)
		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)
	}
	fmt.Printf("[commesso %d] ha finito\n", id)
	done <- true
}

func fornitore(id int) {
	fmt.Printf("[fornitore %d] si è svegliato\n", id)

	var ric Richiesta
	ric.id = id
	ric.tipo = ID_FORNITORE
	ric.ack = make(chan int, MAXBUFF)

	richiestaUscita := 0
	for {
		fmt.Printf("[fornitore %d] e` pronto a fare una consegna\n", id)
		consegnaMascherine <- ric
		richiestaUscita = <-ric.ack // waiting for server ack

		if richiestaUscita == -1 {
			break
		}

		fmt.Printf("[fornitore %d] ha finito la consegna\n", id)

		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)

	}
	fmt.Printf("[fornitore %d] ha finito\n", id)
	done <- true
}

// ritorna l'id di un employee con meno di MAX utenti o -1
func getFreeEmployeeID(utentiDipendenti map[int]int, commessiInServizio map[int]Richiesta) int {
	for id, _ := range commessiInServizio {
		count := 0
		for _, empId := range utentiDipendenti {
			if empId == id {
				count++
			}
			if count >= CLIENTI_MAX_PER_DIPENDENTI {
				break
			}
		}
		if count < CLIENTI_MAX_PER_DIPENDENTI {
			return id
		}
	}
	return -1
}

func canEmployeeExit(id int, utentiDipendenti map[int]int) bool {
	for _, empId := range utentiDipendenti {
		if empId == id {
			return false
		}
	}
	return true
}

// goroutine rappresentante il sistema di gestione del negozio
func server() {
	utentiDipendenti := make(map[int]int)         //mappatura utenti dentro il negozio a dipendenti
	commessiInServizio := make(map[int]Richiesta) //commessi attualmente dentro il negozio, li mappo ad una richiesta in modo da avere il canale ack per eventualmente contattarli durante il servizio
	commessiCheVoglionoUscire := make(map[int]Richiesta)
	mascherineDisponibili := 0
	closing := false

	for {
		//condizioni

		/*
			da gestire
				var ingresso [N_TIPOLOGIE_UTENTI_CHE_ACCEDONO]chan Richiesta
				var ottenimentoMascherina chan Richiesta
				var consegnaMascherine chan Richiesta
				var uscita chan Richiesta
		*/
		select {

		// accesso dipendenti, solo se cap negozio < MAX
		case ric := <-when(closing || (len(commessiInServizio)+len(utentiDipendenti) <= MAX), ingresso[ID_COMMESSO]):
			if closing {
				ric.ack <- -1
			} else {
				commessiInServizio[ric.id] = ric
				ric.ack <- 1
			}

		// accesso utenti abituali, solo se cap negozio < MAX, priorità a dipendenti, ci deve essere un dipendente libero
		case ric := <-when(closing || (len(commessiInServizio)+len(utentiDipendenti) <= MAX && len(ingresso[ID_COMMESSO]) == 0 && getFreeEmployeeID(utentiDipendenti, commessiInServizio) != -1), ingresso[ID_CLIENTE_ABITUALE]):
			//associo l'utente ad un commesso
			utentiDipendenti[ric.id] = getFreeEmployeeID(utentiDipendenti, commessiInServizio)
			ric.ack <- 1

		// accesso utenti non abituali, solo se cap negozio < MAX, priorità a dipendenti e dipendenti abituali, ci deve essere un dipendente libero
		case ric := <-when(len(commessiInServizio)+len(utentiDipendenti) <= MAX && len(ingresso[ID_COMMESSO]) == 0 && len(ingresso[ID_CLIENTE_ABITUALE]) == 0 && getFreeEmployeeID(utentiDipendenti, commessiInServizio) != -1, ingresso[ID_CLIENTE]):
			//associo l'utente ad un commesso
			utentiDipendenti[ric.id] = getFreeEmployeeID(utentiDipendenti, commessiInServizio)
			ric.ack <- utentiDipendenti[ric.id]

		//consegna mascherine a clienti
		case ric := <-when(mascherineDisponibili > 0, ottenimentoMascherina):
			mascherineDisponibili -= 1
			ric.ack <- 1

		// ricezione mascherine da fornitore
		case ric := <-consegnaMascherine:
			if closing {
				ric.ack <- -1
			} else {
				mascherineDisponibili += NM
				ric.ack <- 1
			}

		//uscita
		case ric := <-uscita:
			if ric.tipo == ID_COMMESSO {
				if canEmployeeExit(ric.id, utentiDipendenti) {
					delete(commessiInServizio, ric.id)
					ric.ack <- 1
				} else {
					commessiCheVoglionoUscire[ric.id] = ric
				}
			} else { //cliente
				ric.ack <- 1
				//controllo se il dipendente che seguiva l'utente voleva uscire e se può adesso farlo
				idCommessoCliente := utentiDipendenti[ric.id]
				ricDipendenteDiUscita, vuoleUscire := commessiCheVoglionoUscire[idCommessoCliente]
				if vuoleUscire && canEmployeeExit(idCommessoCliente, utentiDipendenti) {
					ricDipendenteDiUscita.ack <- 1
					delete(commessiCheVoglionoUscire, idCommessoCliente)
					delete(commessiInServizio, idCommessoCliente)
				}
				delete(utentiDipendenti, ric.id)
			}

		case <-closeSignalCmd: // when all users ended
			closing = true
			for _, ricDiUscita := range commessiCheVoglionoUscire {
				ricDiUscita.ack <- 1
			}

		case <-endServerCmd: // when all routines(users+lifeguards) ended
			fmt.Println("THE END !!!!!!")
			done <- true
			return
		}
	}
}

func main() {
	//init canali
	/*
		var ingresso [N_TIPOLOGIE_UTENTI_CHE_ACCEDONO]chan Richiesta
				var ottenimentoMascherina chan Richiesta
				var consegnaMascherine chan Richiesta
				var uscita chan Richiesta
	*/
	for i := 0; i < N_TIPOLOGIE_UTENTI_CHE_ACCEDONO; i++ {
		ingresso[i] = make(chan Richiesta, MAXBUFF)
	}
	ottenimentoMascherina = make(chan Richiesta, MAXBUFF)
	consegnaMascherine = make(chan Richiesta, MAXBUFF)
	uscita = make(chan Richiesta, MAXBUFF)

	go server()
	idProg := 0
	for i := 0; i < N_CLIENTI; i++ {
		go cliente(idProg, rand.Intn(2))
		idProg++
	}
	for i := 0; i < N_DIPENDENTI; i++ {
		go commesso(idProg)
		idProg++
	}
	for i := 0; i < N_FORNITORI; i++ {
		go fornitore(idProg)
		idProg++
	}

	for i := 0; i < N_CLIENTI; i++ {
		<-done
	}
	closeSignalCmd <- true
	for i := 0; i < N_DIPENDENTI+N_FORNITORI; i++ {
		<-done
	}
	endServerCmd <- true
	<-done //waiting for server confirmation

	fmt.Printf("\n HO FINITO ")
}
