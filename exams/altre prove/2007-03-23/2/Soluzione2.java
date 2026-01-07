public class Soluzione2 {

	private static final int N = 8; //numero thread uomo
	private static final int M = 5; //numero thread donna
	
	
	public static void main(String[] args) {
		
	
		
		Uomo lista_u[] = new Uomo[N];
		Donna lista_d[] = new Donna[M];
		
		GestoreAscensore gestore = new GestoreAscensore();
		Cabina C=new Cabina(gestore);
	
	
		for(int i=0;i<N;i++)
		{
			lista_u[i]  = new Uomo(gestore);
			lista_u[i].start();
		}
		
		for(int i=0;i<M;i++)
		{
			lista_d[i]  = new Donna(gestore);
			lista_d[i].start();
		}
		C.start();
		
}	}//fine main