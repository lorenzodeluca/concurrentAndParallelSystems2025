import java.util.Random;


public class noleggio {

	public static void main(String[] args) {
		int N = 100;
		int i, j, k;
		Random r=new Random();
		ClienteFamily[] clientiFamily = new ClienteFamily[N];
		ClienteBusiness[] clientiBusiness = new ClienteBusiness[N];
		Monitor monitor = new Monitor();
		for (i = 0, j=0, k=0; i < N; i++) {
			if(r.nextBoolean())
				clientiFamily[j++] = new ClienteFamily(monitor, r.nextInt()%2);
			else
				clientiBusiness[k++] = new ClienteBusiness(monitor, r.nextInt()%2);
		}
		for (i = 0; i < j; i++) 
			clientiFamily[i].start();
		for (i = 0; i < k; i++) 	
			clientiBusiness[i].start();
		
	}
	
}