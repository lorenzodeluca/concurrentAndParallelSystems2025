/**
 * @(#)dentista.java
 *
 * dentista application
 *
 * @author 
 * @version 1.00 2008/4/14
 */
 


import java.util.Random;


public class dentista {

	public static void main(String[] args) {
		int N = 15;
		int i, j, k, l,  r;
		Random R= new Random();
		paziente_normale[] PN = new paziente_normale[N];
		paziente_ortodontico[] PO = new paziente_ortodontico[N];
		urgente[] PU = new urgente[N];
		monitor m = new monitor();
		
		for (i = 0, j=0, k=0,l=0 ; i < N; i++) {
			r=R.nextInt(); if (r<=0) r=3;
			switch(r%3)
			{	case 0: PU[j++]= new urgente(m); break;
				case 1: PN[k++]= new paziente_normale(m);break;
				case 2: PO[l++]= new paziente_ortodontico(m);
			}
		}
		
		System.out.println("main:ho creato"+j+" pazienti urgenti, "+k+" pazienti normali e "+l+" pazienti ortodontici.");
		for (i = 0; i < j; i++) 
			PU[i].start();
		for (i = 0; i < k; i++) 	
			PN[i].start();
		for (i = 0; i < l; i++) 	
			PO[i].start();
		
	}
	
}