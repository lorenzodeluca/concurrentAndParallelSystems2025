public class ClienteFamily extends Thread {
	
	private Monitor monitor;
	private int veicolo;

	public ClienteFamily(Monitor m, int v) {
		monitor = m;
		veicolo = v;
	}
	
	public void run() {
		try {
			int assegnato;
			sleep((int) (Math.random() * 5000));
			if (veicolo == 0) {
				assegnato = monitor.richiestaFurgoneClienteFamily();
				sleep((int) (Math.random() * 5000));
				if (assegnato == 0)
					monitor.rilascioFurgone();
				else
					monitor.rilascioPromiscuo();
			} else {
				assegnato = monitor.richiestaPullminoClienteFamily();
				sleep((int) (Math.random() * 5000));
				if (assegnato == 0)
					monitor.rilascioPullmino();
				else
					monitor.rilascioPromiscuo();
			}
		} 
		catch (InterruptedException e) {}
	}
	
}
