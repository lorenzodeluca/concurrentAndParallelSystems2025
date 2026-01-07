#include <pthread.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <math.h>
#include <stdint.h>
#include <semaphore.h>
#include <stdbool.h>

#define N 3 // groups arriving at the park
#define MaxP 5 // maximum number of people that can stay inside the park
#define MaxC 2  // number of cars

typedef struct {
	int free_spots;
	int free_cars;
	int order_arrival;
	int gruppi_sospesi[N];
	pthread_mutex_t m; //semaphore to protect the access to this structure
	pthread_cond_t condition;
} park;

park data;

void sleep_ms(int ms) {
	struct timespec ts;
	ts.tv_sec = ms / 1000;
	ts.tv_nsec = (ms % 1000) * 1000000; 
	nanosleep(&ts, NULL);
}

void* parkTour(void *id) {
	int user = (intptr_t) id;
	bool has_ticket = false;
	int groupSize = (1 + rand() % 5);

	printf("\n%d: group size is %d\n", user, groupSize);
	
	pthread_mutex_lock(&data.m);
	int my_order = data.order_arrival++;
	while (!(data.free_spots >= groupSize && data.free_cars > 0 && data.gruppi_sospesi[my_order] == 0)) {
		data.gruppi_sospesi[user]++;
		pthread_cond_wait(&data.condition, &data.m);
		data.gruppi_sospesi[user]--;
	}

	data.free_spots -= groupSize;
	data.free_cars--;
	has_ticket = true;

	pthread_mutex_unlock(&data.m);

	printf("\n%d: got everything, starting to visit the park\n", user);
	sleep_ms(3000);
	printf("\n%d: finished visiting the park\n", user);

	pthread_mutex_lock(&data.m);
	data.free_spots += groupSize;
	data.free_cars++;
	pthread_cond_broadcast(&data.condition);
	pthread_mutex_unlock(&data.m);

	printf("\n%d: end of visit to the park\n", user);
	pthread_exit(NULL);
}

int main() {
	pthread_t users[N];
	int rc;

	pthread_mutex_init(&data.m, NULL);
	pthread_cond_init(&data.condition, NULL);
	data.free_spots = MaxP;
	data.free_cars = MaxC;
	data.order_arrival = 0;
	memset(data.gruppi_sospesi, 0, sizeof(data.gruppi_sospesi));

	for (int t = 0; t < N; t++) {
		rc = pthread_create(&users[t], NULL, parkTour, (void*)(intptr_t)t);
		if (rc) {
			printf("ERROR creating thread %d\n", rc);
			exit(-1);
		}
	}

	for (int t = 0; t < N; t++) {
		rc = pthread_join(users[t], NULL);
		if (rc) {
			printf("ERROR joining thread %d\n", rc);
			exit(-1);
		}
	}

	pthread_mutex_destroy(&data.m);
	pthread_cond_destroy(&data.condition);

	printf("\nFINISHED! :) \n");
	return 0;
}
