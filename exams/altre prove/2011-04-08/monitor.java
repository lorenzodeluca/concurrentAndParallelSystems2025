import java.util.concurrent.locks.*;

public class monitor {

	private final int Nd = 2;//dentisti normali
	private final int No = 2; //ortodonzisti
	private final int M = 5; // poltrone a disposizione dei pazienti
	
	private Lock lock = new ReentrantLock();
	//code per l'attesa delle poltrone:
	private Condition CP_U = lock.newCondition(); // coda per l'acquisizione poltrone dei pazienti urgenti
	private Condition CP_A = lock.newCondition();   // coda per l'acquisizione poltrone di tutti gli altri (i non urgenti)
	private int sospCP_U;
	private int sospCP_A;
	
	//code per l'attesa dei dentisti:
	private Condition CD_N = lock.newCondition(); //coda pazienti normali
	private Condition CD_O = lock.newCondition(); //coda pazienti ortodontici
	private Condition CD_U = lock.newCondition();	// coda pazienti urgenti
	private int sospCD_N;
	private int sospCD_O;
	private int sospCD_U;
	
	private int poltrone_disp; //poltrone disponibili
	private int norm_disp;		//dentisti normali disponibili
	private int orto_disp;		// dentisti ortodonzia disponibili
	private int sospesiFurgoneClienteFamily;

	
	public monitor() {
		poltrone_disp=M; //poltrone disponibili
		norm_disp=Nd;		//dentisti normali disponibili
		orto_disp=No;	
		sospCP_U=0;
		sospCP_A=0;
		sospCD_N=0;
		sospCD_O=0;
		sospCD_U=0;
	}
	
	public void inizioterapiaN() throws InterruptedException { //void: i pazienti normali sono curati sempre da dentisti normali
		lock.lock();
		try {
		// acquisizione di una poltrona:
		while((poltrone_disp==0)||(sospCP_U>0))
		{	sospCP_A++;
			CP_A.await();
			sospCP_A--;
		}
		
		poltrone_disp--;
		stampastato("Paziente normale ha ottenuto una poltrona....");
		//acquisizione del dentista:
		
		while ((norm_disp==0)||(sospCD_U>0))
		{	sospCD_N++;
			CD_N.await();
			sospCD_N--;
		}	
		norm_disp--;
		stampastato("Paziente normale ha ottenuto il dentista! Inizio terapia....");	
		}
		finally {
			lock.unlock();
		}
	}
	
	public void inizioterapiaO() throws InterruptedException { //i pazienti ortodontici sono sempre curati da ortodonzisti
		lock.lock();
		try {
		// acquisizione di una poltrona:
		while((poltrone_disp==0)||(sospCP_U>0))
		{	sospCP_A++;
			CP_A.await();
			sospCP_A--;
		}
		
		poltrone_disp--;
		stampastato("Paziente ortodontico ha ottenuto una poltrona....");
		//acquisizione del dentista:
		
		while (orto_disp==0)
		{	sospCD_O++;
			CD_O.await();
			sospCD_O--;
		}	
		orto_disp--;
		stampastato("Paziente ortodontico ha ottenuto il dentista! Inizio terapia....");		
			
		}
		finally {
			lock.unlock();
		}
	}
	
	public int inizioterapiaU() throws InterruptedException { //int: i pazienti urgenti possono essere curati da normali o ortodonzisti
		lock.lock();
		try {
		int ris;	
		// acquisizione di una poltrona:
		while(poltrone_disp==0)
		{	sospCP_U++;
			CP_U.await();
			sospCP_U--;
		}
		
		poltrone_disp--;
		stampastato("Paziente urgente ha ottenuto una poltrona....");
		//acquisizione del dentista:
		
		while (((orto_disp==0) && (norm_disp==0))||
				((orto_disp>0) && (sospCD_O>0)))
		{	sospCD_U++;
			CD_U.await();
			sospCD_U--;
		}	
		if (norm_disp>0) 
		{	ris=0;
			norm_disp--;
			stampastato("Paziente urgente ha ottenuto un dentista normale! Inizio terapia....");	
		}
		else
		{	ris=1;
			orto_disp--;
			stampastato("Paziente urgente ha ottenuto un ortodonzista! Inizio terapia....");	
		}	
		return ris;
		}
		finally {
			lock.unlock();
		}
	}
	
	
	
	public void fineterapiaN() throws InterruptedException {
		lock.lock();
		try {
			
		
			// rilascio dentista:
			norm_disp++;
			if (sospCD_U>0)
				CD_U.signal();
			else if (sospCD_N>0)
				CD_N.signal();
				stampastato("Fine terapia (normale)- rilasciato dentista..");	
			// rilascio della poltrona:
			poltrone_disp++;
			if (sospCP_U>0)
					CP_U.signal();
			else if 
				(sospCP_A>0)
					CP_A.signal();
			stampastato("Fine terapia(normale)- rilasciata poltrona..");		
			}
		finally {
			lock.unlock();
		}
	}
	
	public void fineterapiaO() throws InterruptedException {
		lock.lock();
		try {
			
			// rilascio dentista:
			orto_disp++;
			if (sospCD_O>0)
				CD_O.signal();
			else if (sospCD_U>0)
				CD_U.signal();
			stampastato("Fine terapia (ortodonzista)- rilasciato dentista..");	
			// rilascio della poltrona:
			poltrone_disp++;
			if (sospCP_U>0)
				CP_U.signal();
			else if (sospCP_A>0)
				CP_A.signal();
				stampastato("Fine terapia(ortodonzista)- rilasciata poltrona..");	
		}
		finally {
			lock.unlock();
		}
	}
	
	void stampastato(String S)
	{ System.out.println(S+"CODE: CP_U="+sospCP_U+"; CP_A="+sospCP_A+"; CD_N="+sospCD_N+"; CD_O="+sospCD_O+"; CD_U="+sospCD_U+". RISORSE: Poltrone="+poltrone_disp+", D_Norm="+norm_disp+", D_orto="+orto_disp);
	}
	
}
