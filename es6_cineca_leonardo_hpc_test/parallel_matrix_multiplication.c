#include <stdio.h>
#include <stdlib.h>
#include <mpi.h>

#define MAX_MATRIX_SIZE 1000

// function to read a matrix from a file
void load_matrix(const char *filename, double matrix[MAX_MATRIX_SIZE][MAX_MATRIX_SIZE], int *rows, int *cols) {
    FILE *file = fopen(filename, "r");
    if (file == NULL) {
        perror("Errore nell'aprire il file");
        exit(EXIT_FAILURE);
    }

    fscanf(file, "%d %d", rows, cols);
    for (int i = 0; i < *rows; i++) {
        for (int j = 0; j < *cols; j++) {
            fscanf(file, "%lf", &matrix[i][j]);
        }
    }

    fclose(file);
}

// function to write a matrix to a file
void write_matrix(const char *filename, double matrix[MAX_MATRIX_SIZE][MAX_MATRIX_SIZE], int rows, int cols) {
    FILE *file = fopen(filename, "w");
    if (file == NULL) {
        perror("Errore nell'aprire il file di output");
        exit(EXIT_FAILURE);
    }

    fprintf(file, "%d %d\n", rows, cols);
    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < cols; j++) {
            fprintf(file, "%.2f ", matrix[i][j]);
        }
        fprintf(file, "\n");
    }

    fclose(file);
}

// function to calculate a single number of the result matrix
double calculate_element(double matrix1[MAX_MATRIX_SIZE][MAX_MATRIX_SIZE], double matrix2[MAX_MATRIX_SIZE][MAX_MATRIX_SIZE],
                         int row, int col, int cols_matrix1) {
    double result = 0.0;
    for (int k = 0; k < cols_matrix1; k++) {
        result += matrix1[row][k] * matrix2[k][col];
    }
    return result;
}

int main(int argc, char *argv[]) {
    int my_rank, comm_sz;
    int rows_matrix1, cols_matrix1, rows_matrix2, cols_matrix2;
    double matrix1[MAX_MATRIX_SIZE][MAX_MATRIX_SIZE], matrix2[MAX_MATRIX_SIZE][MAX_MATRIX_SIZE], result[MAX_MATRIX_SIZE][MAX_MATRIX_SIZE];
    int rows_per_process, start_idx, end_idx;

    MPI_Init(NULL, NULL);
    MPI_Comm_rank(MPI_COMM_WORLD, &my_rank);
    MPI_Comm_size(MPI_COMM_WORLD, &comm_sz);

    if (my_rank == 0) {
        // Carica le matrici dal file CSV
        load_matrix("matrix1.txt", matrix1, &rows_matrix1, &cols_matrix1);
        load_matrix("matrix2.txt", matrix2, &rows_matrix2, &cols_matrix2);
        
        // checking matrix sizes
        if (cols_matrix1 != rows_matrix2) {
            printf("Errore:  matrix1 columns number must be equals matrix2 rows number\n");
            MPI_Abort(MPI_COMM_WORLD, 1);
        }
    }

    // sending matrixes sizes to nodes
    MPI_Bcast(&cols_matrix1, 1, MPI_INT, 0, MPI_COMM_WORLD);
    MPI_Bcast(&rows_matrix2, 1, MPI_INT, 0, MPI_COMM_WORLD);
    MPI_Bcast(&cols_matrix2, 1, MPI_INT, 0, MPI_COMM_WORLD);

    // sending matrixes to nodes
    MPI_Bcast(&matrix1, MAX_MATRIX_SIZE * MAX_MATRIX_SIZE, MPI_DOUBLE, 0, MPI_COMM_WORLD);
    MPI_Bcast(&matrix2, MAX_MATRIX_SIZE * MAX_MATRIX_SIZE, MPI_DOUBLE, 0, MPI_COMM_WORLD);

    // making each node calculate one number of the result matrix
    int total_elements = rows_matrix1 * cols_matrix2;
    int elements_per_process = total_elements / comm_sz; 

    // global index of the number that this node must calculate
    start_idx = my_rank * elements_per_process;
    end_idx = (my_rank == comm_sz - 1) ? total_elements : start_idx + elements_per_process;

    for (int idx = start_idx; idx < end_idx; idx++) {
        int row = idx / cols_matrix2; 
        int col = idx % cols_matrix2; 
        result[row][col] = calculate_element(matrix1, matrix2, row, col, cols_matrix1);
    }

    // the rank 0 node receive all the results and save them in the output file
    if (my_rank != 0) {
        MPI_Send(&result[start_idx / cols_matrix2][start_idx % cols_matrix2], elements_per_process, MPI_DOUBLE, 0, 0, MPI_COMM_WORLD);
    } else {
        for (int source = 1; source < comm_sz; source++) {
            MPI_Recv(&result[source * elements_per_process / cols_matrix2][source * elements_per_process % cols_matrix2],
                     elements_per_process, MPI_DOUBLE, source, 0, MPI_COMM_WORLD, MPI_STATUS_IGNORE);
        }
        write_matrix("matrixResult.txt", result, rows_matrix1, cols_matrix2);
    }

    MPI_Finalize();
    return 0;
}
