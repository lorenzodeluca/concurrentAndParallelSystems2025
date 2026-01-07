#!/bin/bash

#SBATCH --account=tra25_Inginfbo
#SBATCH --partition=g100_usr_prod
#SBATCH --nodes=2
#SBATCH --ntasks-per-node=48
#SBATCH -o job.out
#SBATCH -e job.err
#SBATCH --mail-user=lorenzo@deluca.pro
module load autoload intelmpi
srun ./sol6sr
