/**
 * @(#)paziente_ortodontico.java
 *
 *
 * @author 
 * @version 1.00 2008/4/14
 */


 
public class paziente_ortodontico extends Thread {
	
	private monitor Mon;
	
	public paziente_ortodontico(monitor m) {
		Mon = m;
	
	}
	
	public void run() {
		try {
			
			sleep((int) (Math.random() * 5000));
			Mon.inizioterapiaO();
				sleep((int) (Math.random() * 5000));
			Mon.fineterapiaO();
			
		} 
		catch (InterruptedException e) {}
	}
	
}
