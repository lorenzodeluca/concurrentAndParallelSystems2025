/**
 * @(#)paziente_normale.java
 *
 *
 * @author 
 * @version 1.00 2008/4/14
 */


 
public class paziente_normale extends Thread {
	
	private monitor Mon;
	
	public paziente_normale(monitor m) {
		Mon = m;
	
	}
	
	public void run() {
		try {
			
			sleep((int) (Math.random() * 5000));
			Mon.inizioterapiaN();
				sleep((int) (Math.random() * 5000));
			Mon.fineterapiaN();
			
		} 
		catch (InterruptedException e) {}
	}
	
}
