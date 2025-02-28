package example4.order;

public class PaypalPaymentProcessor extends PaymentProcessor {
    private boolean authenticated;
    private String securityCode;

    public PaypalPaymentProcessor(String securityCode) {
        this.authenticated = false;
        this.securityCode = securityCode;
    }

    @Override
    public void pay(Order order) {
        System.out.println("Processing paypal payment");
        System.out.println("Verifying email: " + this.securityCode);
        order.setStatus("paid");
    }

    @Override
    public void auth_sms(Order order, String securityCode) {
        throw new UnsupportedOperationException("Not implemented");
    }

}
