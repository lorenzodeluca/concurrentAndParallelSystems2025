package main

import (
	"fmt"
	"math/rand"
	"time"
)

const scolari = 25
const N = 40   //numero massimo di persone in sala (compresi sorveglianti)
const NC = 30  // numero massimo di persone nel corridoio
const MaxS = 4 // numero massimo sorveglianti in sala
const MAXBUFF = 15
const MAXPROC = 5
const IN int = 0
const OUT int = 1
const SING int = 0 //visitatori singoli
const SCOL int = 1 //scolaresca di 25 persone
const SORV int = 2 //sorvegliante

type richiesta struct {
	id   int
	tipo int
	ack  chan int
}

var entrataC_IN [3]chan richiesta               //entrata corridoio in direzione IN
var entrataC_OUT [3]chan richiesta              //entrata corridoio in direzione OUT
var uscitaC_IN = make(chan richiesta, MAXBUFF)  //uscita corridoio in direzione IN
var uscitaC_OUT = make(chan richiesta, MAXBUFF) //uscita corridoio in direzione IN

var done = make(chan bool, MAXBUFF)
var termina = make(chan bool, MAXBUFF)

func printTipo(typ int) string {
	switch typ {
	case SING:
		return "visitatore singolo"
	case SCOL:
		return "scolaresca"
	case SORV:
		return "sorvegliante"
	default:
		return ""
	}
}

func printDirezione(typ int) string {
	switch typ {
	case IN:
		return "ingresso"
	case OUT:
		return "uscita"
	default:
		return ""
	}
}

func server() {
	scolaresche_in_C := [2]int{0, 0} // scolaresche nel corridoio per direzione x := [5]int{10, 20, 30, 40, 50}
	persone_in_C := [2]int{0, 0}     // persone nel corridoio per direzione
	var persone_in_sala = 0          //sala vuota all'inizio
	var sorveglianti_in_sala = 0     // nessun sorvegliante

	for {
		select {
		//entrata corridoio IN:
		case x := <-when(scolaresche_in_C[OUT] == 0 &&
			(persone_in_C[IN]+persone_in_C[OUT]) < NC &&
			persone_in_sala < N &&
			sorveglianti_in_sala < MaxS &&
			(len(entrataC_OUT[SCOL])+len(entrataC_OUT[SORV])+len(entrataC_OUT[SING]) == 0), entrataC_IN[SORV]):
			persone_in_C[IN]++
			persone_in_sala++
			sorveglianti_in_sala++
			x.ack <- 1
		case x := <-when(scolaresche_in_C[OUT] == 0 &&
			(persone_in_C[IN]+persone_in_C[OUT]) < NC &&
			persone_in_sala < N &&
			sorveglianti_in_sala > 0 &&
			(len(entrataC_IN[SORV])+len(entrataC_OUT[SCOL])+len(entrataC_OUT[SORV])+len(entrataC_OUT[SING]) == 0), entrataC_IN[SING]):
			persone_in_C[IN]++
			persone_in_sala++
			x.ack <- 1
		case x := <-when(persone_in_C[OUT] == 0 &&
			(persone_in_C[IN]+persone_in_C[OUT])+scolari <= NC &&
			persone_in_sala+scolari <= N &&
			sorveglianti_in_sala > 0 &&
			(len(entrataC_IN[SORV])+len(entrataC_IN[SING])+len(entrataC_OUT[SCOL])+len(entrataC_OUT[SORV])+len(entrataC_OUT[SING]) == 0), entrataC_IN[SCOL]):
			persone_in_C[IN] += scolari
			scolaresche_in_C[IN]++
			persone_in_sala += scolari
			x.ack <- 1
		//entrata corridoio OUT:
		case x := <-when(scolaresche_in_C[IN] == 0 &&
			(persone_in_C[IN]+persone_in_C[OUT]) < NC &&
			(sorveglianti_in_sala > 1 || persone_in_sala == 1) &&
			(len(entrataC_OUT[SCOL])+len(entrataC_OUT[SING]) == 0), entrataC_OUT[SORV]):
			persone_in_C[OUT]++
			persone_in_sala--
			sorveglianti_in_sala--
			x.ack <- 1
		case x := <-when(scolaresche_in_C[IN] == 0 &&
			(persone_in_C[IN]+persone_in_C[OUT]) < NC &&
			(len(entrataC_OUT[SCOL]) == 0), entrataC_OUT[SING]):
			persone_in_C[OUT]++
			persone_in_sala--
			x.ack <- 1
		case x := <-when(persone_in_C[IN] == 0 &&
			(persone_in_C[IN]+persone_in_C[OUT])+scolari <= NC, entrataC_OUT[SCOL]):
			persone_in_C[OUT] += scolari
			scolaresche_in_C[OUT]++
			persone_in_sala -= scolari
			x.ack <- 1
		// uscita dal corridoio
		case x := <-uscitaC_IN:
			if x.tipo == SCOL {
				persone_in_C[IN] -= scolari
				scolaresche_in_C[IN]--
			} else { // singoli o sorveglianti
				persone_in_C[IN]--
			}
			x.ack <- 1
		case x := <-uscitaC_OUT:
			if x.tipo == SCOL {
				persone_in_C[OUT] -= scolari
				scolaresche_in_C[OUT]--
			} else { // singoli o sorveglianti
				persone_in_C[OUT]--
			}
			x.ack <- 1
		case <-termina: // quando tutti i processi hanno finito
			fmt.Println("\nFINE !!!!!!")
			done <- true
			return
		}
	}
}

func visitatore(id int, tipo int) {
	var tt int
	var r richiesta
	tt = rand.Intn(2) + 1
	fmt.Printf("\nInizializzazione visitatore %d del tipo %s  in secondi %d \n", id, printTipo(tipo), tt)
	time.Sleep(time.Duration(tt) * time.Second)
	r = richiesta{id, tipo, make(chan int, MAXBUFF)}
	entrataC_IN[tipo] <- r // send asincrona
	<-r.ack                // attesa x sincronizzazione
	fmt.Printf("\n[visitatore %d di tipo %s]  entro nel corridoio in direzione IN \n", id, printTipo(tipo))
	tt = rand.Intn(2) + 1
	time.Sleep(time.Duration(tt) * time.Second)
	uscitaC_IN <- r
	<-r.ack
	fmt.Printf("\n[visitatore %d di tipo %s]  entrato in sala \n", id, printTipo(tipo))
	tt = rand.Intn(5) + 1 // tempo di visita
	time.Sleep(time.Duration(tt) * time.Second)
	entrataC_OUT[tipo] <- r // send asincrona
	<-r.ack                 // attesa x sincronizzazione
	fmt.Printf("\n[visitatore %d di tipo %s]  entro nel corridoio in direzione OUT \n", id, printTipo(tipo))
	tt = rand.Intn(2) + 1
	time.Sleep(time.Duration(tt) * time.Second)
	uscitaC_OUT <- r
	<-r.ack
	fmt.Printf("\n[visitatore %d di tipo %s]  esco dal corridoio in direzione OUT e vado a casa...\n", id, printTipo(tipo))
	done <- true
}

func sorvegliante(id int) {
	var tt int
	var r richiesta
	var tipo = SORV
	volte := 2 * MAXPROC // per assicurare che ci sia lameno un sorvegliante in esecuzione per ogni possibile visitatore
	tt = rand.Intn(2) + 1
	fmt.Printf("\nInizializzazione sorvegliante %d in secondi %d ...\n", id, tt)
	time.Sleep(time.Duration(tt) * time.Second)
	r = richiesta{id, tipo, make(chan int, MAXBUFF)}

	for i := 0; i < volte; i++ {
		entrataC_IN[tipo] <- r // send asincrona
		<-r.ack                // attesa x sincronizzazione
		fmt.Printf("\n[sorvegliante %d ]  entro nel corridoio in direzione IN \n", id)
		tt = rand.Intn(2) + 1
		time.Sleep(time.Duration(tt) * time.Second)
		uscitaC_IN <- r
		<-r.ack
		fmt.Printf("\n[sorvegliante %d ]  presidio la sala \n", id)
		tt = rand.Intn(5) + 1 // tempo di visita
		time.Sleep(time.Duration(tt) * time.Second)
		entrataC_OUT[tipo] <- r // send asincrona
		<-r.ack                 // attesa x sincronizzazione
		fmt.Printf("\n[sorvegliante %d ]  entro nel corridoio in direzione OUT \n", id)
		tt = rand.Intn(2) + 1
		time.Sleep(time.Duration(tt) * time.Second)
		uscitaC_OUT <- r
		<-r.ack
		fmt.Printf("\n[sorvegliante %d]  esco dal corridoio in direzione OUT...\n", id)
		tt = rand.Intn(1) + 1
		time.Sleep(time.Duration(tt) * time.Second)
	}
	fmt.Printf("\n[sorvegliante %d]  ho finito e vado a casa...\n", id)
	done <- true
}
func when(b bool, c chan richiesta) chan richiesta {
	if !b {
		return nil
	}
	return c
}

func main() {
	var scolaresche int
	var singoli int
	var sorveglianti int
	rand.Seed(time.Now().Unix())
	fmt.Printf("\n[main] quante scolaresche?(max %d)", MAXPROC)
	fmt.Scanf("%d", &scolaresche)
	fmt.Printf("\n[main] quanti singoli?(max %d)", MAXPROC)
	fmt.Scanf("%d", &singoli)
	fmt.Printf("\n[main] quanti sorveglianti?(max %d)", MAXPROC)
	fmt.Scanf("%d", &sorveglianti)
	//inizializzazione canali clienti
	for i := 0; i < 3; i++ {
		entrataC_IN[i] = make(chan richiesta, MAXBUFF)
		entrataC_OUT[i] = make(chan richiesta, MAXBUFF)
	}

	go server()
	for i := 0; i < sorveglianti; i++ {
		go sorvegliante(i)
	}
	for i := 0; i < singoli; i++ {
		go visitatore(i, SING)
	}
	for i := 0; i < scolaresche; i++ {
		go visitatore(i, SCOL)
	}

	for i := 0; i < (sorveglianti + singoli + scolaresche); i++ {
		<-done
		fmt.Printf("\n[main] terminato %d-simo processo...\n\n", i+1)
	}
	fmt.Printf("\n[main] finiti gli utenti\n\n")
	termina <- true // terminazione server
	<-done
}
