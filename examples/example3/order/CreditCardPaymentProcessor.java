package example3.order;

public class CreditCardPaymentProcessor implements PaymentProcessor {

    @Override
    public void pay(Order order, String securityCode) {
        System.out.println("Processing credit card payment");
        System.out.println("Verifying security code: " + securityCode);
        order.setStatus("paid");
    }
}
