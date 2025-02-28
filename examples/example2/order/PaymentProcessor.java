package example2.order;

public class PaymentProcessor {

    public void pay(Order order, String securityCode) {
        System.out.println("Processing payment");
        System.out.println("Verifying security code: " + securityCode);
        order.setStatus("paid");
    }
}
