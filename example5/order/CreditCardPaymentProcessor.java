package example5.order;

public class CreditCardPaymentProcessor extends PaymentProcessor {

    private String securityCode;

    public CreditCardPaymentProcessor(String securityCode) {
        this.securityCode = securityCode;
    }

    @Override
    public void pay(Order order) {
        System.out.println("Processing debit card payment");
        System.out.println("Verifying security code: " + this.securityCode);
        order.setStatus("paid");
    }
}
