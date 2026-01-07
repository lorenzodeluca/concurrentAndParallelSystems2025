// appello 11 gennaio 2010 - tema D


#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <time.h>
#define N1 3			// numero go kart C1
#define N2 3			// numero go kart C2
#define NUM_MI 5			// num thread minorenni
#define NUM_MA 5			//num thread maggiorenni

#define C1 0				// Tipo go kart C1
#define C2 1				// Tipo go kart C2

int verboso; 

typedef struct
{	pthread_mutex_t lock;
	pthread_cond_t codaMA; // coda maggiorenni	
	pthread_cond_t codaMI; // coda minorenni 
	int sospMA; // clienti sospesi MA
	int sospMI; // clienti sospesi MI
	int inPista[2];	// go kart in pista
	int disp[2]; // go kart disponibili
} pista;

pista Pista;

void debug(pista *p)
{
	printf("\nSTATO DELLA PISTA:\n");
	printf("Go kart C1 in pista: %d\n", p->inPista[C1]);
	printf("Go kart C2 in pista: %d\n", p->inPista[C2]);
	printf("MA in attesa: %d\n", p->sospMA);
	printf("MI in attesa: %d\n", p->sospMI);
	printf("C1 disponibili: %d\n", p->disp[C1]);
	printf("C2 disponibili: %d\n", p->disp[C2]);
	printf("\n");
}

int InPistaMA(pista *p) //utilizzano C1 e C2
{	
	int ris=C1;
	pthread_mutex_lock (&p->lock);
	
	while( p->disp[C1]==0 && p->disp[C2]==0 ||     //Non ci sono kart disponibili
		(p->disp[C1]>0 && p->inPista[C2]>0 )|| // potrebbe acquisire un C1, ma ci sono C2 in pista
		(p->disp[C1]==0 && p->disp[C2]>0 && p->inPista[C1]>0)) //potrebbe acquisire un C2, ma ci sono C1 in pista
	{
		printf("MA sospeso in ingresso\n");
		p->sospMA++;
		pthread_cond_wait(&p->codaMA, &p->lock);
		p->sospMA--;
	}
	printf("MA entra \n");

	if(p->disp[C1]>0)
	{
		p->disp[C1]--;
		p->inPista[C1]++;
		ris=C1;
	}
	else
	{	p->disp[C2]--;
		p->inPista[C2]++;
		ris= C2;
	}
	
	if (verboso) debug(p); // per debug
	
	pthread_mutex_unlock(&p->lock);
	return ris;
}

void InPistaMI(pista *p) //utilizzano solo C1
{	
	pthread_mutex_lock (&p->lock);
	
	while( 	p->disp[C2]==0 || /*Non ci sono C2 disponibili*/
		p->sospMA>0 || /*Ci sono degli Ma che attendono*/
		p->inPista[C1]>0)  /*ci sono  in pista dei C1*/
	{
		printf("MI sospeso in ingresso\n"); 
		p->sospMI++;
		pthread_cond_wait(&p->codaMI, &p->lock);
		p->sospMI--;
	}
	printf("MI entra\n");	
	p->disp[C2]--;
	p->inPista[C2]++;
	if (verboso) debug(p);
	pthread_mutex_unlock(&p->lock);
}

void OutPista(pista *p, int tipo) // possono aver utilizzato sia C1 che C2
{	
	pthread_mutex_lock (&p->lock);
	printf("Rilascio go kart di tipo %d\n", tipo);	
	p->inPista[tipo]--;
	p->disp[tipo]++;
	switch (tipo) 
	{	
	
	case C1:	if (p->inPista[tipo]>0) // pista non vuota: entra un altro thread maggiorenne
				{	if (p->sospMI>0) 
						pthread_cond_signal(&p->codaMA); //
				}
				else if (p->inPista[tipo]==0) // pista vuota: -> potrebbero entrare maggiorenni e minorenni con C2
				{	if (p->sospMA>0) 
						pthread_cond_broadcast(&p->codaMA);
					if(p->sospMI>0) 
						pthread_cond_broadcast(&p->codaMI);
				}
				break;
	case C2:	if (p->inPista[tipo]>0) // pista non vuota: entra un altro thread
				{	if (p->sospMA>0) 
						pthread_cond_signal(&p->codaMA); //
					else if (p->sospMI>0) 
						pthread_cond_signal(&p->codaMI);
				}
				else	// pista vuota: -> potrebbero entrare i maggiorenni  con  C1
				{	if (p->sospMA>0) 
						pthread_cond_broadcast(&p->codaMA);
				}
				break;
	}
	if (verboso) debug(p);
	pthread_mutex_unlock(&p->lock);
}



void *MA(void * arg)
{   
	int tipo;
	printf("MA attivato.. \n");
	tipo=InPistaMA(&Pista);
	sleep(3); /* simulazione uso risorsa */
	OutPista(&Pista, tipo);
}

void *MI(void * arg)
{   	
	printf("MI attivato.. \n");
	InPistaMI(&Pista);
	sleep(3); 
	OutPista(&Pista, C2);
}

void init_pista(pista *p)
{ 
	pthread_mutex_init(&p->lock, NULL);
	pthread_cond_init(&p->codaMA, NULL);
	pthread_cond_init(&p->codaMI, NULL);
	p->sospMA = 0;
	p->sospMI = 0;
	p->inPista[C1] = 0;
	p->inPista[C2] = 0;
	p->disp[C1]=N1;
	p->disp[C2]=N2;
}

int main ()
{
	int i;
	pthread_t clienti_ma[NUM_MA];
	pthread_t clienti_mi[NUM_MI];
	printf("esecuzione verbosa? [0/1]\n");
	scanf("%d", &verboso);
	init_pista(&Pista);

	
	for(i = 0; i < NUM_MI; i++)
		pthread_create(&clienti_mi[i], NULL, MI, NULL);
		
		
	for(i = 0; i < NUM_MA; i++)
		pthread_create(&clienti_ma[i], NULL, MA, NULL);

	for(i = 0; i < NUM_MA; i++)
		pthread_join(clienti_ma[i], NULL);
	for(i = 0; i < NUM_MI; i++)
		pthread_join(clienti_mi[i], NULL);

	return 0;
}
