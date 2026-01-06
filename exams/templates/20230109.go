package main

import (
	"fmt"
	"math/rand"
	"time"
)

// limiti
const NS = 2 // posti standard auto
const NM = 1 // posti maxi usabili per auto e camper

// info simulazione
const NUMERO_AUTO = 3
const NUMERO_CAMPER = 3
const NUMERO_SPAZZANEVE = 1
const TEMPO_ATTESA_MAX = 5 //tempo per simulare i tempi di transito/permanenza

// altre variabili
const MAXBUFF = 100
const MAXPROC = NUMERO_AUTO + NUMERO_CAMPER + NUMERO_SPAZZANEVE //per sapere quanti proc in esecuzione oltre al server
const NUM_DIREZIONI = 2
const NUMERO_PARCHEGGI = NS + NM + NUMERO_SPAZZANEVE
const TIPO_AUTO = 0
const TIPO_CAMPER = 1
const TIPO_SPAZZANEVE = 2
const DIREZIONE_ENTRATA = 0
const DIREZIONE_USCITA = 1

// risorse/canali
var transitoAuto [NUM_DIREZIONI]chan Richiesta
var transitoCamper [NUM_DIREZIONI]chan Richiesta
var transitoSpazzaneve [NUM_DIREZIONI]chan Richiesta
var notificaUscitaStrada [NUM_DIREZIONI]chan Richiesta

var closeSignalCmd = make(chan bool)
var endServerCmd = make(chan bool)
var done = make(chan bool)

// classi/strutture utili
type Richiesta struct {
	id   int //id utente
	tipo int //tipo parcheggio richiesto, 0=STANDARD/AUTO, 1=MAXI/AUTO/CAMPER, 2=spazzaneve
	dir  int // DIREZIONE_ENTRATA = 0, DIREZIONE_USCITA = 1
	ack  chan int
}

type Parcheggio struct {
	idParcheggio int
	idUtente     int
	isMaxi       bool
	isSpazzaneve bool
}

func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

func tipoUtenteToString(tipo int) string {
	switch tipo {
	case TIPO_AUTO:
		return "AUTO"
	case TIPO_CAMPER:
		return "CAMPER"
	default:
		return "SPAZZANEVE"
	}
}

func user(id int, tipo int) { //auto e camper
	fmt.Printf("[user %d/%s] 1. è pronto a salire la strada\n", id, tipoUtenteToString(tipo))

	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.dir = DIREZIONE_ENTRATA
	ric.ack = make(chan int, MAXBUFF)

	//accede alla Strada in direzione Salita
	if tipo == TIPO_AUTO {
		transitoAuto[DIREZIONE_ENTRATA] <- ric
	} else {
		transitoCamper[DIREZIONE_ENTRATA] <- ric
	}
	<-ric.ack // waiting for server ack

	fmt.Printf("[user %d/%s] 2. percorre la strada in salita impiegando un tempo arbitrario \n", id, tipoUtenteToString(tipo))
	time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)
	fmt.Printf("[user %d/%s] 3.esce dalla strada in direzione Salita occupando un parcheggio nel piazzale \n", id, tipoUtenteToString(tipo))

	//notifica fine transito strada e occupazione parcheggio
	notificaUscitaStrada[DIREZIONE_ENTRATA] <- ric
	idParcheggio := <-ric.ack // waiting for server ack
	fmt.Printf("[user %d/%s] 4. visita del castello in un tempo arbitrario, si è parcheggiato al parcheggio numero %d \n", id, tipoUtenteToString(tipo), idParcheggio)
	time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second) //4. visita del castello in un tempo arbitrario

	fmt.Printf("[user %d/%s] 5. pronto a partire \n", id, tipoUtenteToString(tipo))
	ric.dir = DIREZIONE_USCITA
	if tipo == TIPO_AUTO {
		transitoAuto[DIREZIONE_USCITA] <- ric
	} else {
		transitoCamper[DIREZIONE_USCITA] <- ric
	}
	<-ric.ack // waiting for server ack
	fmt.Printf("[user %d/%s] 6. percorre la strada in discesa impiegando un tempo arbitrario \n", id, tipoUtenteToString(tipo))
	time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)
	notificaUscitaStrada[DIREZIONE_ENTRATA] <- ric
	<-ric.ack // waiting for server ack
	fmt.Printf("[user %d/%s] 7. esce dalla strada in direzione Discesa\n", id, tipoUtenteToString(tipo))
	done <- true
}

func spazzaneve(id int) {
	tipo := TIPO_SPAZZANEVE
	fmt.Printf("[spazzaneve %d] avviato\n", id)

	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.ack = make(chan int, MAXBUFF)

	richiestaUscita := 0
	for {
		fmt.Printf("[spazzaneve %d] 1. accede alla strada in direzione Discesa\n", id)
		ric.dir = DIREZIONE_USCITA
		transitoSpazzaneve[DIREZIONE_USCITA] <- ric
		richiestaUscita = <-ric.ack // waiting for server ack

		if richiestaUscita == -1 {
			break
		}

		fmt.Printf("[user %d/%s] 2. percorre la strada in discesa impiegando un tempo arbitrario \n", id, tipoUtenteToString(tipo))
		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)
		fmt.Printf("[user %d/%s] 3.esce dalla strada in direzione discesa  \n", id, tipoUtenteToString(tipo))

		//notifica fine transito strada e occupazione parcheggio
		notificaUscitaStrada[DIREZIONE_USCITA] <- ric
		<-ric.ack // waiting for server ack
		fmt.Printf("[user %d/%s] 4. si ubriaca al bar\n", id, tipoUtenteToString(tipo))
		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)

		fmt.Printf("[user %d/%s] 5. accede alla strada in direzione Salita \n", id, tipoUtenteToString(tipo))
		ric.dir = DIREZIONE_ENTRATA
		transitoSpazzaneve[DIREZIONE_ENTRATA] <- ric
		<-ric.ack // waiting for server ack
		fmt.Printf("[user %d/%s] 6. percorre la strada in salita impiegando un tempo arbitrario \n", id, tipoUtenteToString(tipo))
		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)

		fmt.Printf("[user %d/%s] 7. esce dalla strada in direzione salita\n", id, tipoUtenteToString(tipo))
		notificaUscitaStrada[DIREZIONE_ENTRATA] <- ric
		<-ric.ack // waiting for server ack
		fmt.Printf("[user %d/%s] rimane nel piazzale per un tempo arbitrario\n", id, tipoUtenteToString(tipo))
		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)

	}
	fmt.Printf("[spazzaneve %d] ha finito\n", id)
	done <- true
}

// ritorna userId oppure -1
func presenteTipoInDirezione(dir int, tipo int, inTransito map[int]Richiesta) int {
	for userId, info := range inTransito {
		if info.dir == dir && info.tipo == tipo {
			return userId
		}
	}
	return -1
}

// ritorna id di un parcheggio standard(solo auto) libero
func trovaParcheggioStandardLibero(parcheggi [NUMERO_PARCHEGGI]Parcheggio) int {
	for i := 0; i < NS+NM; i++ {
		if parcheggi[i].idUtente == -1 && parcheggi[i].isMaxi == false && parcheggi[i].isSpazzaneve == false {
			return i
		}
	}
	return -1
}

// ritorna id di un parcheggio maxi(auto o camper) libero
func trovaParcheggioMaxiLibero(parcheggi [NUMERO_PARCHEGGI]Parcheggio) int {
	for i := 0; i < NS+NM; i++ {
		if parcheggi[i].idUtente == -1 && parcheggi[i].isMaxi == true && parcheggi[i].isSpazzaneve == false {
			return i
		}
	}
	return -1
}

// goroutine used to control the access to the center and its areas
func server() {
	var parcheggi [NUMERO_PARCHEGGI]Parcheggio
	inTransito := make(map[int]Richiesta) //utenti attualmente in transito
	closing := false

	//inizializzando le variabili del server
	for i := 0; i < NS; i++ {
		parcheggi[i] = Parcheggio{idParcheggio: i, idUtente: -1, isMaxi: false, isSpazzaneve: false}
	}
	for i := NS; i < NS+NM; i++ {
		parcheggi[i] = Parcheggio{idParcheggio: i, idUtente: -1, isMaxi: true, isSpazzaneve: false}
	}
	for i := NS + NM; i < NS+NM+NUMERO_SPAZZANEVE; i++ {
		parcheggi[i] = Parcheggio{idParcheggio: i, idUtente: -1, isMaxi: false, isSpazzaneve: true}
	}

	for {
		//condizioni
		inTransitoSpazzaneveInEntrataOUscita := presenteTipoInDirezione(DIREZIONE_ENTRATA, TIPO_SPAZZANEVE, inTransito) != -1 || presenteTipoInDirezione(DIREZIONE_USCITA, TIPO_SPAZZANEVE, inTransito) != -1
		inTransitoCamperInUscita := presenteTipoInDirezione(DIREZIONE_USCITA, TIPO_CAMPER, inTransito) != -1
		inTransitoCamperInEntrata := presenteTipoInDirezione(DIREZIONE_ENTRATA, TIPO_CAMPER, inTransito) != -1
		inTransitoAutoOCamper := presenteTipoInDirezione(DIREZIONE_ENTRATA, TIPO_CAMPER, inTransito) != -1 || presenteTipoInDirezione(DIREZIONE_USCITA, TIPO_CAMPER, inTransito) != -1 || presenteTipoInDirezione(DIREZIONE_ENTRATA, TIPO_AUTO, inTransito) != -1 || presenteTipoInDirezione(DIREZIONE_USCITA, TIPO_AUTO, inTransito) != -1
		inCodaVeicoliPerDiscesa := len(transitoSpazzaneve[DIREZIONE_USCITA]) > 0 || len(transitoCamper[DIREZIONE_USCITA]) > 0 || len(transitoAuto[DIREZIONE_USCITA]) > 0
		//inCodaSpazzanevePerEntrata := len(transitoSpazzaneve[DIREZIONE_ENTRATA]) > 0
		inCodaCamperPerEntrata := len(transitoCamper[DIREZIONE_ENTRATA]) > 0
		//inCodaAutoPerUscita := len(transitoAuto[DIREZIONE_USCITA]) > 0
		inCodaAutoPerEntrata := len(transitoAuto[DIREZIONE_ENTRATA]) > 0
		inCodaCamperPerUscita := len(transitoCamper[DIREZIONE_USCITA]) > 0
		inCodaSpazzanevePerUscita := len(transitoSpazzaneve[DIREZIONE_USCITA]) > 0

		//fmt.Printf("in coda salita: auto:%t, camion:%t, spazzaneve:%t \n", inCodaAutoPerEntrata, inCodaCamperPerEntrata, inCodaSpazzanevePerEntrata)
		//fmt.Printf("in coda discesa: auto:%t, camion:%t, spazzaneve:%t \n", inCodaAutoPerUscita, inCodaCamperPerUscita, inCodaSpazzanevePerUscita)

		select {
		// ingresso/salita auto, non ci devono essere spazzaneve in transito, non ci devono essere camper in senso opposto, bisogna dare priorità a camper, bisogna dare priorità a chi vuole scendere, ci deve essere parcheggio libero
		case ric := <-when(!inTransitoSpazzaneveInEntrataOUscita && !inTransitoCamperInUscita && !inCodaCamperPerEntrata && !inCodaVeicoliPerDiscesa && (trovaParcheggioStandardLibero(parcheggi) != -1 || trovaParcheggioMaxiLibero(parcheggi) != -1), transitoAuto[DIREZIONE_ENTRATA]):
			//trovo il parcheggio per l'auto
			idParcheggio := trovaParcheggioStandardLibero(parcheggi)
			if idParcheggio == -1 {
				idParcheggio = trovaParcheggioMaxiLibero(parcheggi)
			}
			//riservo il parcheggio per l'auto
			parcheggi[idParcheggio].idUtente = ric.id
			//indico l'auto come in transito
			inTransito[ric.id] = ric
			//autorizzo l'auto a salire
			ric.ack <- 1

		//uscita/discesa auto, non ci deve essere lo spazzaneve in transito, non ci devono essere camper in transito nel senso opposto, priorità a camper e spazzaneve in coda in uscita
		case ric := <-when(!inTransitoSpazzaneveInEntrataOUscita && !inTransitoCamperInEntrata && !inCodaCamperPerUscita && !inCodaSpazzanevePerUscita, transitoAuto[DIREZIONE_USCITA]):
			//libero il parcheggio
			idPark := 0
			for i := 0; i < NUMERO_PARCHEGGI; i++ {
				if parcheggi[i].idUtente == ric.id {
					idPark = parcheggi[i].idParcheggio
					break
				}
			}
			parcheggi[idPark].idUtente = -1
			//indico l'auto come in transito
			inTransito[ric.id] = ric
			//autorizzo il transito
			ric.ack <- 1

		// ingresso/salita camper, non ci devono essere spazzaneve in transito,non ci devono essere camper in transito nel senso opposto, priorità code discesa, ci deve essere park maxi libero
		case ric := <-when(!inTransitoSpazzaneveInEntrataOUscita && !inTransitoCamperInUscita && !inCodaVeicoliPerDiscesa && (trovaParcheggioMaxiLibero(parcheggi) != -1), transitoCamper[DIREZIONE_ENTRATA]):
			//trovo il parcheggio per il camper
			idParcheggio := trovaParcheggioMaxiLibero(parcheggi)
			//riservo il parcheggio
			parcheggi[idParcheggio].idUtente = ric.id
			//indico il veicolo come in transito
			inTransito[ric.id] = ric
			//autorizzo il passaggio
			ric.ack <- 1

		//uscita/discesa camper, non ci deve essere lo spazzaneve in transito, non ci devono essere camper in transito nel senso opposto, priorità a spazzaneve in coda in uscita
		case ric := <-when(!inTransitoSpazzaneveInEntrataOUscita && !inTransitoCamperInEntrata && !inCodaSpazzanevePerUscita, transitoCamper[DIREZIONE_USCITA]):
			//libero il parcheggio
			idPark := 0
			for i := 0; i < NUMERO_PARCHEGGI; i++ {
				if parcheggi[i].idUtente == ric.id {
					idPark = parcheggi[i].idParcheggio
					break
				}
			}
			parcheggi[idPark].idUtente = -1
			//indico l'utente come in transito
			inTransito[ric.id] = ric
			//autorizzo il transito
			ric.ack <- 1

		// ingresso/salita spazzaneve,non ci devono essere camper in transito nel senso opposto, priorità a camper e auto, priorità code discesa, ci deve essere park maxi libero
		case ric := <-when(!inTransitoAutoOCamper && !inTransitoCamperInUscita && !inCodaVeicoliPerDiscesa && !inCodaCamperPerEntrata && !inCodaAutoPerEntrata, transitoSpazzaneve[DIREZIONE_ENTRATA]):
			//indico il veicolo come in transito
			inTransito[ric.id] = ric
			//autorizzo il passaggio
			ric.ack <- 1

		//uscita/discesa spazzaneve,  non ci devono essere camper in transito nel senso opposto, priorità a spazzaneve in coda in uscita
		case ric := <-when(!inTransitoAutoOCamper && !inTransitoCamperInEntrata, transitoSpazzaneve[DIREZIONE_USCITA]):
			if closing {
				ric.ack <- -1
			} else {
				inTransito[ric.id] = ric
				//autorizzo il transito
				ric.ack <- 1
			}

		case ric := <-notificaUscitaStrada[DIREZIONE_ENTRATA]:
			//tolgo dal transito
			delete(inTransito, ric.id)
			//notifico il numero di parcheggio già riservato
			idPark := 0
			for i := 0; i < NS+NM; i++ {
				if parcheggi[i].idUtente == ric.id {
					idPark = parcheggi[i].idParcheggio
					break
				}
			}
			ric.ack <- idPark

		case ric := <-notificaUscitaStrada[DIREZIONE_USCITA]:
			//tolgo dal transito
			delete(inTransito, ric.id)
			ric.ack <- 1

		case <-closeSignalCmd: // when all users ended
			closing = true

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
		var transitoAuto [NUM_DIREZIONI]chan Richiesta
		var transitoCamper [NUM_DIREZIONI]chan Richiesta
		var transitoSpazzaneve [NUM_DIREZIONI]chan Richiesta
		var notificaUscitaStrada [NUM_DIREZIONI]chan Richiesta
	*/
	for i := 0; i < NUM_DIREZIONI; i++ {
		transitoAuto[i] = make(chan Richiesta, MAXBUFF)
		transitoCamper[i] = make(chan Richiesta, MAXBUFF)
		transitoSpazzaneve[i] = make(chan Richiesta, MAXBUFF)
		notificaUscitaStrada[i] = make(chan Richiesta, MAXBUFF)
	}

	go server()
	idProg := 0
	for i := 0; i < NUMERO_AUTO; i++ {
		go user(idProg, TIPO_AUTO)
		idProg++
	}
	for i := 0; i < NUMERO_CAMPER; i++ {
		go user(idProg, TIPO_CAMPER)
		idProg++
	}
	for i := 0; i < NUMERO_SPAZZANEVE; i++ {
		go spazzaneve(idProg)
		idProg++
	}

	for i := 0; i < NUMERO_AUTO+NUMERO_CAMPER; i++ {
		<-done
	}
	closeSignalCmd <- true
	for i := 0; i < NUMERO_SPAZZANEVE; i++ {
		<-done
	}
	endServerCmd <- true //telling server to shut down
	<-done               //waiting for server confirmation

	fmt.Printf("\n HO FINITO ")
}
