public class Donna extends Thread{

	private GestoreAscensore gestore;
	
	public Donna(GestoreAscensore g)
	{
		gestore = g;
	}
	
	public void run()
	{
			try
			{				
					System.out.println("Creata donna");
					gestore.donnaSale();
					System.out.println("	Donna  arrivata  in cima alla torre");
					
					sleep(3000);
																		
					gestore.donnaScende();
					System.out.println("	Donna : arrivata al piano terra");
										
					
					
									
					
				
			}//fine try
			catch(Exception ex)
			{
				System.out.println("-- Errore catturato nel thread donna --");
				ex.printStackTrace();
				System.out.println("----------------------------------");
			}
	
	}//fine run

}
