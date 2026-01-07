/**
 * @(#)urgente.java
 *
 *
 * @author 
 * @version 1.00 2008/4/14
 */


public class urgente extends Thread {
	
	private  monitor Mon;

	public urgente(monitor m) {
		Mon = m;
		
	}
	
	public void run() {
		try {
			int assegnato;
			sleep((int) (Math.random() * 5000));
			assegnato = Mon.inizioterapiaU();
			sleep((int) (Math.random() * 5000));
			if (assegnato == 0) // dentista normale
					Mon.fineterapiaN();
				else
					Mon.fineterapiaO();
		} 
		catch (InterruptedException e) {}
	}
	
}
