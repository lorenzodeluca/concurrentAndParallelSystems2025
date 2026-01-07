//aluigi andrea 0000272395 compito 4

import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

public class GestoreAscensore{

	private final int MAXA = 10;	//num massimo di persone nell'ascensore
	private final int MAXT = 15;	//num massimo di persone nella cima della torre
		
	private int numPersoneA;			//num persone contenute nell'ascensore attualmente
	private int numPersoneTorre;		//num persone contenute nella cima della torre attualmente
	
	private boolean ascensoreTerra;		//a true se l'ascensore si trova a terra
	private boolean ascensoreCima;		//a true se l'ascensore si trova in cima
	
	//lock per la mutua esclusione fra i metodi del gestore
	private Lock lock = new ReentrantLock();
	
	//variabili condizione:
	private Condition TerraUomo = lock.newCondition();	//coda per gli uomini a terra
	private Condition TerraDonna = lock.newCondition();//coda per le donne a terra
	private Condition CimaUomo = lock.newCondition();	//coda per gli uomini in cima alla torre
	private Condition CimaDonna = lock.newCondition();	//coda per le donne in cima alla torre
	private Condition ascT=lock.newCondition();
	private Condition ascC=lock.newCondition();
	private Condition codaArrivo=lock.newCondition(); //coda di uscita dall'ascensore
	private int sospTU;
	private int sospTD;
	private int sospCU;
	private int sospCD;
	private boolean inviaggio;
	
	public GestoreAscensore()
	{
		//inizializzazioni variabili
		numPersoneA = 0;		//ascensore inizialmente vuoto
		numPersoneTorre = 0;	//cima della torre inizialmente vuota
		inviaggio=false;		
		ascensoreTerra = true;	//l'ascensore si trova gia pronto a terra all'arrivo dei threads
		ascensoreCima = false;
		sospTU=0; sospTD=0;sospCU=0; sospCD=0;
	
	}
	
	
	
	public void uomoSale() throws InterruptedException
	{
		lock.lock();
		try
		{
			
			while( (ascensoreTerra == false) ||  (numPersoneA== MAXA) || ((numPersoneA+numPersoneTorre)==MAXT) ||
					sospTD>0)
			{	sospTU++;
				System.out.println("uomo stop terra");
				TerraUomo.await();
				sospTU--;
			}
			
			System.out.println("uomo entrato in ascensore");
			numPersoneA++;
			
			// l'ascensore e` pronto per salire se:		
			if (prontoSalita())
			{	ascT.signal(); //risveglio l'ascensore	
			
			}
			
			while (!ascensoreCima) //attesa di arrivare a destinazione
				codaArrivo.await();
			
			numPersoneA--;
			numPersoneTorre++;
			
			if (numPersoneA==0) //segnalo i processi in attesa di scendere
			{	inviaggio=false;
				CimaDonna.signalAll();
				CimaUomo.signalAll();
			}
												
		}
		
		finally
		{
			lock.unlock();
		}
		return;
	}
	
	
	public void donnaSale() throws InterruptedException
	{
		lock.lock();
		try
		{
			
			while( (ascensoreTerra == false) ||  (numPersoneA>= MAXA) || ((numPersoneA+numPersoneTorre)>=MAXT))
			{	sospTD++;
				System.out.println("donna stop terra - sospesi T: "+sospTD+" "+sospTU+" sospesi C: "+sospCD+" "+ sospCU+
				"persone in A "+numPersoneA+ " persone in cima: "+numPersoneTorre +" asc.a terra: "+ ascensoreTerra);
				
				TerraDonna.await();
				sospTD--;
			}
			
			System.out.println("donna entrata in ascensore");
			numPersoneA++;
			
					
			// l'ascensore e` pronto per salire se:		
			if (prontoSalita())
			{
				ascT.signal(); //risveglio l'ascensore	
			
			}
			while (!ascensoreCima) //attesa di arrivare a destinazione
				codaArrivo.await();
			numPersoneA--;
			numPersoneTorre++;
			
			if (numPersoneA==0) //segnalo i processi in attesa di scendere
			{	inviaggio=false;
				CimaDonna.signalAll();
				CimaUomo.signalAll();
			}
												
		}
		
		finally
		{
			lock.unlock();
		}
		return;
	}
	
	public void uomoScende() throws InterruptedException
	{
		lock.lock();
		try
		{
			
			while( (ascensoreCima == false) ||  (numPersoneA>= MAXA) ||	sospCD>0)
			{	sospCU++;
				System.out.println("uomo stop cima");
				CimaUomo.await();
				sospCU--;
			}
			
			
			numPersoneA++;
			numPersoneTorre--;
					
			if( prontoDiscesa() ) //l'ascensore e` pronto per scendere
			{
				ascC.signal(); //risveglio l'ascensore	
			
			}
			while (!ascensoreTerra) //attesa di arrivare a destinazione
				codaArrivo.await();
			numPersoneA--;
			
			
			if (numPersoneA==0) //segnalo i processi in attesa di salire
			{	inviaggio=false;
				TerraDonna.signalAll();
				TerraUomo.signalAll();
			}
												
		}
		
		finally
		{
			lock.unlock();
		}
		return;
	}
	
	public void donnaScende() throws InterruptedException
	{
		lock.lock();
		try
		{
			
			while( (ascensoreCima == false) ||  (numPersoneA>= MAXA) )
			{	sospCD++;
				System.out.println("donna stop cima");
				CimaDonna.await();
				sospCD--;
			}
			
			
			numPersoneA++;
			numPersoneTorre--;
					
			if( prontoDiscesa() ) //l'ascensore e` pronto per scendere
			{	ascC.signal(); //risveglio l'ascensore	
			
			}
			while (!ascensoreTerra) //attesa di arrivare a destinazione
				codaArrivo.await();
			numPersoneA--;
			
			
			if (numPersoneA==0) //segnalo i processi in attesa di salire
			{	inviaggio=false;
			    TerraDonna.signalAll();
				TerraUomo.signalAll();
			}
												
		}
		
		finally
		{
			lock.unlock();
		}
		return;
	}
	
	
	
	
	public void attendiPT()throws InterruptedException
	{	lock.lock();
		try
		{
			ascensoreTerra=true;
			
			codaArrivo.signalAll(); //escono i passeggeri dalla cabina
			
			while (inviaggio||!prontoSalita()) //attesa riempimento cabina
			{		System.out.println("ascensore stop terra - inviaggio: "+inviaggio+" prontosalita:"+prontoSalita());
						ascT.await();
			}
			inviaggio=true;
			ascensoreTerra=false; //partenza per la cima
					
		}
		
		finally
		{
			lock.unlock();
		}
		return;
	}
	
	public void attendiP1()throws InterruptedException
	{	lock.lock();
		try
		{	ascensoreCima=true;
			
			codaArrivo.signalAll(); //escono i passeggeri dalla cabina
			
			while (inviaggio||(!prontoDiscesa())) //attesa riempimento cabina
			{System.out.println("ascensore stop cima - inviaggio: "+inviaggio+" prontodiscesa:"+prontoDiscesa());
					ascC.await();
			}
			ascensoreCima=false; //partenza per il piano terra
			inviaggio=true;
					
		}
		
		finally
		{
			lock.unlock();
		}
		return;
	}
	
private boolean prontoSalita() //verifica se l'ascensore e` pronto per salire
{ 	if ((numPersoneA==MAXA)|| //l'ascensore e` pieno
		(numPersoneA+numPersoneTorre==MAXT)|| //la torre e` piena
		(sospTD+sospTU==0)) //nessuno e` in coda
		return true;
	else return false;
}	

private boolean prontoDiscesa() //verifica se l'ascensore e` pronto per scendere
{ 	if ((numPersoneA==MAXA)|| //l'ascensore e` pieno
		(sospCD+sospCU==0)) //nessuno e` in coda
		return true;
	else return false;
}	

}//fine gestore
