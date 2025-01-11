package example5.order;

public class PaypalPaymentProcessor extends PaymentProcessor {
    private String securityCode;

    public PaypalPaymentProcessor(String securityCode) {
        this.securityCode = securityCode;
    }

    @Override
    public void pay(Order order) {
        System.out.println("Processing debit card payment");
        System.out.println("Verifying security code: " + this.securityCode);
        order.setStatus("paid");
    }

}
