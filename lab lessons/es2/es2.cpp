#include <pthread.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <math.h>
#include <stdint.h>
#include <semaphore.h>

#define N 3
#define MaxP 1 // numero posti massimo di persone che possono stare dentro il parco
#define MaxB 1  // numero bici
#define MaxM 1  // numero monopattini
#define V sem_wait
#define P sem_post

sem_t S; //semaforo condizione
typedef struct {
	int posti_liberi;
	int bici_libere;
	int monop_liberi;
} parco;

parco data;

void sleep_ms(int ms) {
	struct timespec ts;
	ts.tv_sec = ms / 1000;
	ts.tv_nsec = (ms % 1000) * 1000000;  // ms → ns
	nanosleep(&ts, NULL);
}

void* giroAlParco(void *id) {
	int utente = (intptr_t) id;
	printf("\n%d: in coda alla biglietteria\n", utente);
	bool has_ticket = false, has_vehicle = false;
	bool vuole_bici = (1 + rand() % 10) % 2 == 0;  // Voto casuale tra 1 e 10
	printf("\n%d: vuole bici %d\n", utente, vuole_bici);
	//semaforo di attesa per evitare accessi multipli a struttura dati

	while (!has_ticket && !has_vehicle) {
		sem_wait(&S);
		printf("\n%d: parlando con il commesso\n", utente);
		if (!has_ticket && data.posti_liberi > 0) {
			data.posti_liberi--;
			printf("\n%d: ottenuto biglietto\n", utente);
			has_ticket = true;
		}
		if (has_ticket && !has_vehicle && vuole_bici && data.bici_libere > 0) {
			data.bici_libere--;
			printf("\n%d: ottenuto bici\n", utente);
			has_vehicle = true;
		}
		if (has_ticket && !has_vehicle && !vuole_bici
				&& data.monop_liberi > 0) {
			data.monop_liberi--;
			printf("\n%d: ottenuto monopattino\n", utente);
			has_vehicle = true;
		}
		sem_post(&S);
		sleep_ms(3000);
	};

	printf("\n%d: ottenuto tutto, inizio a visitare il parco\n", utente);
	sleep_ms(3000);
	printf("\n%d: ho finito di visitare il parco\n", utente);

	printf(
			"\n%d: aspetto di poter parlare con il commesso per riconsegnare il mezzo\n",
			utente);
	while (has_vehicle || has_ticket) {
		sem_wait(&S);
		printf("\n%d: parlando con il commesso\n", utente);
		if (has_ticket) {
			data.posti_liberi++;
			printf("\n%d: liberato posto nel parco\n", utente);
			has_ticket = false;
		}
		if (has_vehicle && vuole_bici) {
			data.bici_libere++;
			has_vehicle = false;
			printf("\n%d: restituita bici\n", utente);
		}
		if (has_vehicle && !vuole_bici) {
			data.monop_liberi++;
			has_vehicle = false;
			printf("\n%d: restituito monopattino\n", utente);
		}
		sem_post(&S);
	};
	printf("\n%d: fine della visita al parco\n", utente);
	pthread_exit(NULL);  // Termina il thread
}

int main() {
	pthread_t utenti[N];  // N utenti (threads)
	int rc;
	int status;

	// Inizializzazione dell'array questionari con -1 (voto non assegnato)
	data.posti_liberi = MaxP;
	data.bici_libere = MaxB;
	data.monop_liberi = MaxM;
	sem_init(&S, 0, 1);

	// Creazione dei thread (utenti)
	for (int t = 0; t < N; t++) {
		rc = pthread_create(&utenti[t], NULL, giroAlParco,
				(void*) (intptr_t) t);
		if (rc) {
			printf("ERRORE nella creazione del thread %d\n", rc);
			exit(-1);
		}
	}

	// Attendere che tutti i thread terminino
	for (int t = 0; t < N; t++) {
		rc = pthread_join(utenti[t], (void*) &status);
		if (rc) {
			printf("ERRORE nel join del thread %d\n", rc);
			exit(-1);
		}
	}

	printf("\nFINITO!:\n");
	return 0;
}
