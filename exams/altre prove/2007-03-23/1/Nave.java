import java.util.concurrent.locks.*;


public class Nave {
	
	private final int MAX_A=15;
	private final int MAX_B=15;
	private final int AB=0;
	private final int BA=1;
	
	private Lock lock = new ReentrantLock();
	private Condition codaAO = lock.newCondition();
	private Condition codaBO = lock.newCondition();
	private Condition codaAE = lock.newCondition();
	private Condition codaBE = lock.newCondition();

	private int numA, numB, sospAO, sospBO, sospAE, sospBE,inScala;
	int dirS; //0: A->B; 1: B->A

	
	public Nave(int persA, int persB) 
	{
		numA=persA; numB=persB;
		inScala=0;
		dirS=AB;
		sospAO=0;
		sospBO=0;
		sospAE=0;
		sospBE=0;
		
	}
	
	

	public void InOspite_daA() throws InterruptedException {
		lock.lock();
		try {
			
			while ((dirS==BA && inScala!=0) || (numB+inScala==MAX_B)) {
				sospAO++;
				codaAO.await();
				sospAO--;
			}
			dirS=AB;
			numA--;
			inScala++;
			System.out.println("ospite nella scala che sta andando nel ponte B ");	
		} finally {
			lock.unlock();
		}
	}

public void InOspite_daB() throws InterruptedException {
		lock.lock();
		try {
			
			while ((dirS==AB && inScala!=0) || (numA+inScala==MAX_A)) {
				sospBO++;
				codaBO.await();
				sospBO--;
			}
			dirS=BA;
			numB--;
			inScala++;
			System.out.println("ospite nella scala che sta andando nel ponte A ");	
		} finally {
			lock.unlock();
		}
	}
	
	public void InEquip_daA() throws InterruptedException {
		lock.lock();
		try {
			
			while ((dirS==BA && inScala!=0) || (numB+inScala==MAX_B) || (sospBO>0)&&(!pienoA())|| sospAO>0) {
				sospAE++;
				codaAE.await();
				sospAE--;
			}
			dirS=AB;
			numA--;
			inScala++;
			System.out.println("equipaggio nella scala che sta andando nel ponte B ");	
		} finally {
			lock.unlock();
		}
	}
public void InEquip_daB() throws InterruptedException {
		lock.lock();
		try {
			
			while ((dirS==AB && inScala!=0) || (numA+inScala==MAX_A) || ((sospAO>0)&&(!pienoB()))||sospBO>0) {
				sospBE++;
				codaBE.await();
				sospBE--;
			}
			dirS=BA;
			numB--;
			inScala++;
			System.out.println("equipaggio nella scala che sta andando nel ponte A ");	
		} finally {
			lock.unlock();
		}
	}
	

public void Out_inB() throws InterruptedException {
		lock.lock();
		try {
			
			inScala--;
			numB++;
			if (inScala==0)
			{	codaBO.signalAll();
				codaBE.signalAll();
			}
			
			System.out.println("persona che arriva al ponteB ");	
		} finally {
			lock.unlock();
		}
	}
		
	public void Out_inA() throws InterruptedException {
		lock.lock();
		try {
			
			inScala--;
			numA++;
			if (inScala==0)
			{	codaAO.signalAll();
				codaAE.signalAll();
			}
			
			System.out.println("persona che arriva al ponteA ");	
		} finally {
			lock.unlock();
		}
	}	
	private boolean pienoB()
	{	boolean ret=false;
		if((dirS==AB)&&(inScala>0))
		{	if (numB+inScala==MAX_B) ret=true;
		}
		else
			if (dirS==BA)
				if (numB==MAX_B)	ret=true;
		return ret;
	}
	
	private boolean pienoA()
	{	boolean ret=false;
		if((dirS==BA)&&(inScala>0))
		{	if (numA+inScala==MAX_A) ret=true;
		}
		else
			if (dirS==AB)
				if (numA==MAX_A)	ret=true;
		return ret;
	}
}