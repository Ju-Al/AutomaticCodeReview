package example3.order;

public class PaypalPaymentProcessor implements PaymentProcessor {

    @Override
    public void pay(Order order, String securityCode) {
        System.out.println("Processing paypal payment");
        System.out.println("Verifying email: " + securityCode);
        order.setStatus("paid");
    }

}
