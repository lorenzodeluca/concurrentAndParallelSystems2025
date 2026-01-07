#include <pthread.h>
#include <semaphore.h>
#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#define N 10
#define MAX 3

typedef struct{
int counterP;
sem_t mux, primi, secondi, liberi, fase2;
} mensa;

mensa M;

void init(mensa *m) {
	m->counterP=0; // numero di primi consumati
	sem_init(&m->mux,0,1); //mutua esclusione
	sem_init(&m->primi,0,0); //semaforo risorsa: numero di primi sul bancone
	sem_init(&m->secondi,0,0);//semaforo risorsa: numero di secondi sul bancone
	sem_init(&m->liberi,0,  MAX); //semaforo risorsa: numero di posti disponibili sul bancone
	sem_init(&m->fase2,0,0); // semaforo evento: passaggio alla seconda fase del cuoco
	printf("BENVENUTI ALLA MENSA AZIENDALE!!\n\n");
}



void *dipendente(void *t) // codice spettatore
{  long tid, result=0;
   tid = (int)t;
   printf("dipendente %ld è partito...\n",tid);
	// acquisizione primo piatto
	sleep(1);
   sem_wait(&M.primi); // prelevo un primo
   sem_wait(&M.mux);
   printf("dipendente %ld sta mangiando il primo...\n",tid);
   M.counterP++; //numero di dipendenti che hanno preso il primo
	sem_post(&M.liberi); // libero un posto
	if (M.counterP==N)
	{	printf("l'ultimo dipendente ha consumato il primo!\n\n");
		sem_post(&M.fase2);
	}
   sem_post(&M.mux);
   sleep(1); 
  // passa al secondo
   sem_wait(&M.secondi); //prelevo un secondo
   printf("dipendente %ld sta mangiando il secondo...\n",tid);
   sem_post(&M.liberi);//libero un posto
   pthread_exit((void*) result);
}

void *cuoco(void * p) // codice worker
{  int i;
    
   printf("cuoco è partito...\n");
   for(i=0; i<N; i++)
   {	//sleep(1);
	   sem_wait(&M.liberi);
	   sem_post(&M.primi); // deposito un primo
	}
	sem_wait(&M.fase2); //attesa che tutti i dip abbiano terminato di prelevare i primi	 
	for(i=0; i<N; i++)
	{	//sleep(1);
		sem_wait(&M.liberi);
		sem_post(&M.secondi); // deposito un secondo
	}
	printf("\n\n Il cuoco ha finito...\n");
   pthread_exit(NULL);
}


int main (int argc, char *argv[])
{  pthread_t op, thread[N];
	int rc;
   long t;
   void *status;
   
   init(&M);
  
	for(t=0; t<N; t++) {
      printf("Main: creazione thread %ld\n", t);
      rc = pthread_create(&thread[t], NULL, dipendente, (void *)t); // creazione dipendente
      if (rc) {
         printf("ERRORE: %d\n", rc);
         exit(-1);   }
  }
  pthread_create(&op, NULL, cuoco, NULL); // creazione cuoco
	for(t=0; t<N; t++) {
      rc = pthread_join(thread[t], &status);
      if (rc) 
		   printf("ERRORE join thread &d codice %ld\n", t, rc);
		else 
		printf("Finito thread %ld con ris. %ld\n",t,(long)status);
  }
	
	

  pthread_join(op, &status);
  return 0;
}
