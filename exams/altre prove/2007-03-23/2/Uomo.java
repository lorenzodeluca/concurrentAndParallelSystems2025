//aluigi andrea 0000272395 compito 4

public class Uomo extends Thread{

	private GestoreAscensore gestore;
	
	public Uomo(GestoreAscensore g)
	{
		gestore = g;
	}

	public void run()
	{
			try
			{
			
				System.out.println("Creato uomo");
				gestore.uomoSale();
				System.out.println("	Uomo : in cima alla torre");
				
				sleep(3000);
																	
				gestore.uomoScende();
				System.out.println("	Uomo : sono in cima alla torre");
									
				
					
				
			}//fine try
			catch(Exception ex)
			{
				System.out.println("-- Errore catturato nel thread uomo --");
				ex.printStackTrace();
				System.out.println("----------------------------------");
			}
	
	}//fine run

}
