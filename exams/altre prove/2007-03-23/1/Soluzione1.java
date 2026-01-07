
public class Soluzione1 {
    
  	public static void main(String[] args) {
		
		int i;
		int inAO=7;
		int inBO=7;
		int inAE=7;
		int inBE=7;
		
		int Nospiti, Nequip;
		
		Nave g=new Nave(inAO+inAE,inBO+inBE);
		Nospiti=inAO+inBO;
		Nequip=inAE+inBE;
		Ospite [] ospiti = new Ospite[Nospiti];
		Equipaggio [] equipaggio = new Equipaggio[Nequip];
		
		for (i=0;i<inAO;i++)
			ospiti[i] = new Ospite(0,g);
		
		for (i=inAO; i<Nospiti; i++)
			ospiti[i] = new Ospite(1,g);
			
		for (i=0;i<inAE;i++)
			equipaggio[i] = new Equipaggio(0,g);
		
		for(i=inAE; i<Nequip; i++)
			equipaggio[i] = new Equipaggio(1,g);
		
		for (i=0;i<Nospiti;i++)
			ospiti[i].start();
		for (i=0;i<Nequip;i++)	
			equipaggio[i].start();
		
	
		

	}

}

