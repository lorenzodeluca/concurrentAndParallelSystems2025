#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <time.h>

#define NUM_E 2			// numero gruppi esperti
#define NUM_P 3			// numero gruppi principianti
#define NI 2			// numero di istruttori
#define P 0				// Tipo principianti
#define E 1				// Tipo esperti
#define MAX 30			// capacitÃ  pista

typedef struct{	
	pthread_mutex_t lock;
	pthread_cond_t codaIngresso[2][MAX]; // code di ingresso 
	pthread_cond_t codaUscita[MAX]; 		// code di uscita 
	int sospIngresso[2][MAX]; 			// gruppi sospesi in ingresso principianti ed esperti
	int sospUscita[MAX]; 				// gruppi sospesi in uscita principianti
	int inPista[2];								// persone in pista, principianti ed esperti			
	int istruttori;								// istruttori disponibili
} pista;

pista Pista;


int piu_prioritari_P_IN(pista *p, int num)
{	int i;
	int ris=0;
	for (i=0;i<num-1; i++)
		if (Pista.sospIngresso[P][i]>0)
		{	ris=1;
			break;
		}
	return ris;
}



int piu_prioritari_E_IN(pista *p, int num)
{	int i;
	int ris=0;
	
	if (piu_prioritari_P_IN(p, MAX))
		ris=1;
	else
		for (i=0; i<num-1; i++)
		if (p->sospIngresso[E][i]>0)
		{	ris=1;
			break;
		}
	return ris;
}

int piu_prioritari_P_OUT(pista *p, int num)
{	int ris=0;
	int i;
	for (i=num-1; i< MAX-1; i++)
		if (p->sospUscita[i]>0)
		{	ris=1;
			break;
		}
	return ris;
}


void segnalaP_OUT(pista *p )
{	int i;
	for (i=MAX-1; i>=0; i--)
		pthread_cond_broadcast(&p->codaUscita[i]);

}


void segnalaE_IN(pista *p)
{	int i;
	for (i=0; i<MAX; i++)
		pthread_cond_broadcast(&p->codaIngresso[E][i]);

}


void segnalaP_IN(pista *p)
{	int i;
	for (i=0; i<MAX; i++)
		pthread_cond_broadcast(&p->codaIngresso[P][i]);

}

void InPistaP(pista *p, int num) // num: numero membri del gruppo;
{	
	pthread_mutex_lock (&p->lock);
	
	while(p->inPista[P] + p->inPista[E] + num > MAX || p->istruttori == 0 || piu_prioritari_P_IN(p, num))
		{  	   
			printf("P %d sospeso in ingresso\n", num);
			p->sospIngresso[P][num - 1]++;
			pthread_cond_wait(&p->codaIngresso[P][num - 1], &p->lock);
			p->sospIngresso[P][num - 1]--;
		}
	printf("P %d entra \n", num);
	p->istruttori--;
	p->inPista[P] += num;
	segnalaP_OUT(p);
	segnalaE_IN(p);
	pthread_mutex_unlock(&p->lock);
}


void InPistaE(pista *p, int num) // num: numero membri del gruppo;
{	
	pthread_mutex_lock (&p->lock);
	
	while(	(p->inPista[P] + p->inPista[E] + num > MAX) || 
			(p->inPista[P] < p->inPista[E] + num) ||
			piu_prioritari_E_IN(p, num))
		{  	   
			printf("E %d sospeso in ingresso\n", num);
			p->sospIngresso[E][num - 1]++;
			pthread_cond_wait(&p->codaIngresso[E][num - 1], &p->lock);
			p->sospIngresso[E][num - 1]--;
		}
	printf("E %d entra\n", num);
	p->inPista[E] += num;
	pthread_mutex_unlock(&p->lock);
}


void OutPistaP(pista *p, int num)
{	
	pthread_mutex_lock (&p->lock);
	while (	(p->inPista[P] - num < p->inPista[E] )	|| 
			(piu_prioritari_P_OUT(p, num)))
		{
			printf("P %d sospeso in uscita\n", num);
			p->sospUscita[num - 1]++;
			pthread_cond_wait(&p->codaUscita[num - 1], &p->lock);
			p->sospUscita[num - 1]--;
		}
		printf("P %d esce\n", num);
		p->istruttori++;
		p->inPista[P] -= num;
		
		
		segnalaP_IN(p);
		segnalaE_IN(p);
		pthread_mutex_unlock(&p->lock);
}

void OutPistaE(pista *p, int num)
{	
	pthread_mutex_lock (&p->lock);
	printf("E %d esce\n", num);
	p->inPista[E] -= num;
	segnalaP_OUT(p);
	segnalaP_IN(p);
	segnalaE_IN(p);
	pthread_mutex_unlock(&p->lock);
}


void *gruppoEsperto(void * arg)
{   
	int num = (int)arg;
	
	sleep((int)(rand() % 3 + 1));
	printf("E %d attivato.. \n", num);
	InPistaE(&Pista, num);
	sleep(3); /* simulazione uso risorsa */
	OutPistaE(&Pista, num);
}

void *gruppoPrincipiante(void * arg)
{   	
	int num = (int)arg;
		
	sleep((int)(rand() % 3 + 1));
	printf("P %d attivato.. \n", num);
	InPistaP(&Pista, num);
	sleep(3); /* simulazione uso risorsa */
	OutPistaP(&Pista, num);
}

void init_pista(pista *p)
{ 
	int i; 
	
	pthread_mutex_init(&p->lock, NULL);
	
	for(i = 0; i < MAX; i++)
	{
		pthread_cond_init(&p->codaIngresso[P][i], NULL);
		pthread_cond_init(&p->codaIngresso[E][i], NULL);
		pthread_cond_init(&p->codaUscita[i], NULL);
		p->sospIngresso[P][i] = 0;
		p->sospIngresso[E][i] = 0;
		p->sospUscita[i] = 0;
	}
	
	p->istruttori = NI;
	p->inPista[P] = 0;
	p->inPista[E] = 0;
}

int main ()
{
	int i;
	pthread_t esperti[NUM_E];
	pthread_t principianti[NUM_P];

	init_pista(&Pista);
	srand(time(NULL));

	for(i = 0; i < NUM_P; i++)
		pthread_create(&principianti[i], NULL, gruppoPrincipiante, (void*)(rand() % MAX + 1));
	for(i = 0; i < NUM_E; i++)
		pthread_create(&esperti[i], NULL, gruppoEsperto, (void*)(rand() % MAX + 1));

	for(i = 0; i < NUM_E; i++)
		pthread_join(esperti[i], NULL);
	for(i = 0; i < NUM_P; i++)
		pthread_join(principianti[i], NULL);

	return 0;
}
