public class Cabina extends Thread{

	private GestoreAscensore gestore;
	
	public Cabina(GestoreAscensore g)
	{
		gestore = g;
	}
	
	public void run()
	{	int i;
			try
			{		System.out.println("Creata cabina");
					for(i=0; i<50; i++)		
					{	gestore.attendiPT();
						System.out.println("Ascensore: sto salendo...");
						sleep(3000);
						gestore.attendiP1();
						System.out.println("	Ascensore: sto scendendo...");
					}
			}//fine try
			catch(Exception ex)
			{
				System.out.println("-- Errore catturato nel thread donna --");
				ex.printStackTrace();
				System.out.println("----------------------------------");
			}
	
	}//fine run

}