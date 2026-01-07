import java.util.concurrent.locks.*;

public class Monitor {

	private final int N_FURGONI = 10;
	private final int N_PULLMINI = 12;
	private final int N_PROMISCUI = 5;
	
	private Lock lock = new ReentrantLock();
	private Condition codaFurgoneClienteFamily = lock.newCondition();
	private Condition codaPullminoClienteFamily = lock.newCondition();
	private Condition codaFurgoneClienteBusiness = lock.newCondition();
	private Condition codaPullminoClienteBusiness = lock.newCondition();
	
	private int furgoni;
	private int pullmini;
	private int promiscui;
	private int sospesiFurgoneClienteFamily;
	private int sospesiPullminoClienteFamily;
	private int sospesiFurgoneClienteBusiness;
	private int sospesiPullminoClienteBusiness;
	
	public Monitor() {
		furgoni = N_FURGONI;
		pullmini = N_PULLMINI;
		promiscui = N_PROMISCUI;
		sospesiFurgoneClienteFamily = 0;
		sospesiPullminoClienteFamily = 0;
		sospesiFurgoneClienteBusiness = 0;
		sospesiPullminoClienteBusiness = 0;
	}
	
	public int richiestaFurgoneClienteFamily() throws InterruptedException {
		lock.lock();
		try {
			while ((furgoni == 0 && promiscui == 0)||(sospesiFurgoneClienteBusiness>0))  
			{	sospesiFurgoneClienteFamily++;
				codaFurgoneClienteFamily.await();
				sospesiFurgoneClienteFamily--;
			}
			
			if (furgoni > 0) {
				furgoni--;
				System.out.println("Furgone assegnato a cliente family - In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 0;
			}
			else {
				promiscui--;
				System.out.println("Mezzo promiscuo assegnato a cliente family(FF)- In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 1;
			}
		}
		finally {
			lock.unlock();
		}
	}
	
	public int richiestaPullminoClienteFamily() throws InterruptedException {
		lock.lock();
		try {
			while (	((pullmini == 0 && promiscui == 0)||(sospesiPullminoClienteBusiness>0))||
					((promiscui>0)&&(sospesiFurgoneClienteBusiness>0))||
					((promiscui>0)&&(sospesiFurgoneClienteFamily>0)) )
			{
				sospesiPullminoClienteFamily++;
				codaPullminoClienteFamily.await();
				sospesiPullminoClienteFamily--;
			}
			if (pullmini >0) {
				pullmini--;
				System.out.println("Pullmino assegnato a cliente family- In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 0;
			}
			else {
				promiscui--;
				System.out.println("Mezzo promiscuo assegnato a cliente family (FP)- In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 1;
			}
		}
		finally {
			lock.unlock();
		}
	}
	
	public int richiestaFurgoneClienteBusiness() throws InterruptedException {
		lock.lock();
		try {
			while (furgoni == 0 && promiscui == 0) {
				sospesiFurgoneClienteBusiness++;
				codaFurgoneClienteBusiness.await();
				sospesiFurgoneClienteBusiness--;
			}
			if (furgoni >0) {
				furgoni--;
				System.out.println("Furgone assegnato a cliente business - In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 0;
			}
			else {
				promiscui--;
				System.out.println("Mezzo promiscuo assegnato a cliente business(BF)- In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 1;
			}
		}
		finally {
			lock.unlock();
		}
	}
	
	public int richiestaPullminoClienteBusiness() throws InterruptedException {
		lock.lock();
		try {
			while ((pullmini == 0 && promiscui == 0) ||
					((promiscui>0)&&(sospesiFurgoneClienteBusiness>0))||
					((promiscui>0)&&(sospesiFurgoneClienteFamily>0)) )
			{
				sospesiPullminoClienteBusiness++;
				codaPullminoClienteBusiness.await();
				sospesiPullminoClienteBusiness--;
			}
			if (pullmini >0 ) {
				pullmini--;
				System.out.println("Pullmino assegnato a cliente business- In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 0;
			}
			else {
				promiscui--;
				System.out.println("Mezzo promiscuo assegnato a cliente business (P) - In attesa: BF="+sospesiFurgoneClienteBusiness+ "\tBP="+sospesiPullminoClienteBusiness
					+ "\tFF="+sospesiFurgoneClienteFamily + "\tFP="+sospesiPullminoClienteFamily);
				return 1;
			}
		}
		finally {
			lock.unlock();
		}
	}
	
	public void rilascioFurgone() throws InterruptedException {
		lock.lock();
		try {
			furgoni++;
			if (sospesiFurgoneClienteBusiness > 0)
				codaFurgoneClienteBusiness.signal();
			else if (sospesiFurgoneClienteFamily > 0)
				codaFurgoneClienteFamily.signal();
			System.out.println("Rilasciato furgone");
		}
		finally {
			lock.unlock();
		}
	}
	
	public void rilascioPullmino() throws InterruptedException {
		lock.lock();
		try {
			pullmini++;
			if (sospesiPullminoClienteBusiness > 0)
				codaPullminoClienteBusiness.signal();
			else if (sospesiPullminoClienteFamily > 0)
				codaPullminoClienteFamily.signal();
			System.out.println("Rilasciato pullmino");
		}
		finally {
			lock.unlock();
		}
	}
	
	public void rilascioPromiscuo() throws InterruptedException {
		lock.lock();
		try {
			promiscui++;
			if (sospesiFurgoneClienteBusiness > 0)
				codaFurgoneClienteBusiness.signal();
			else if (sospesiFurgoneClienteFamily > 0)
				codaFurgoneClienteFamily.signal();
			else if (sospesiPullminoClienteBusiness > 0)
				codaPullminoClienteBusiness.signal();
			else if (sospesiPullminoClienteFamily > 0)
				codaPullminoClienteFamily.signal();
			System.out.println("Rilasciato mezzo promiscuo");
		}
		finally {
			lock.unlock();
		}
	}
	
}
