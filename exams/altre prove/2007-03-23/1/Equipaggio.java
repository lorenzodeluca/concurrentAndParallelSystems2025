public class Equipaggio extends Thread{

	Nave g;
	int p;

	public Equipaggio(int P, Nave G){
		p=P;
		g=G;
	
	}
	
	public void run() {
		try {
			for (int i = 0; i < 4; i++) {
				if (p == 0) {
					sleep((int)(Math.random()*5000));
					g.InEquip_daA();
					sleep((int)(Math.random()*1000));
					g.Out_inB();
					p=1;
				} else {
					sleep((int)(Math.random()*5000));
					g.InEquip_daB();
					sleep((int)(Math.random()*1000));
					g.Out_inA();
					p=0;
				}
			}
		} catch (Exception e) {}
		 
		
	}
}