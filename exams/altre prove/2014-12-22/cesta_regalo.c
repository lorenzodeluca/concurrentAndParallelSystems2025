#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <time.h>
#include <math.h>
#include <unistd.h>
#include <semaphore.h>

#define N 3

#define THREAD N+1

sem_t daPreparare[N];
sem_t preparati[N];
int cestiPronti;

void confeziona()
{
	int i;
	for(i=0;i<N;i++)
	{
		sem_wait(&preparati[i]);
	}
    
	cestiPronti++;
    
	for(i=0;i<N;i++)
	{
		sem_post(&daPreparare[i]);
	}
    
}
void * produttore(void *arg)
{
	int id;
	id = *((int*)arg);
	while(1)
	{   sem_wait(&daPreparare[id]);
        printf("[produttore %d] aggiunto articolo:%d\n",id,id);
        sem_post(&preparati[id]);
		sleep(1);
	}
	pthread_exit(NULL);
	return NULL;
}
void * confezionatore(void *arg)
{
	int id, i;
	id = *((int*)arg);
	while(1)
	{	for(i=0;i<N;i++) //attesa completamento articoli
        {
            sem_wait(&preparati[i]);
        }
        cestiPronti++;
        printf("[confezionatore %d] finita cesta n. %d\n",id,cestiPronti);
        sleep(3);
        for(i=0;i<N;i++) //riattivazione produttori
        {
            sem_post(&daPreparare[i]);
        }
		sleep(2);
	}
	pthread_exit(NULL);
	return NULL;
}


void init()
{
	int i;
	for(i=0;i<N;i++)
	{
		sem_init(&preparati[i],0,0);
		sem_init(&daPreparare[i],0,1);
	}
	cestiPronti=0;
}

int main()
{
	int i,ids[THREAD];
	pthread_t thread[THREAD];
	//pthread_t altroThread;
	i=0;
	init();
	for(i=0;i<N;i++)
	{
		ids[i]=i;
		pthread_create(&thread[i],NULL,produttore,&ids[i]);
	}
	
	ids[N]=i;
	pthread_create(&thread[i],NULL,confezionatore,&i);
	sleep(THREAD*3);
	for(i=0;i<THREAD;i++)
	{
		pthread_join(thread[i],NULL);
	}
	return 0;
}
