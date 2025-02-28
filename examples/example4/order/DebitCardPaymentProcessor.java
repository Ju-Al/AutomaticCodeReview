package example4.order;

public class DebitCardPaymentProcessor extends PaymentProcessor {
    private boolean authenticated;
    private String securityCode;

    public DebitCardPaymentProcessor(String securityCode) {
        this.authenticated = false;
        this.securityCode = securityCode;
    }

    @Override
    public void pay(Order order) {
        if (!this.authenticated) {
            throw new IllegalStateException("Not authentificated");
        }
        System.out.println("Processing debit card payment");
        System.out.println("Verifying security code: " + securityCode);
        order.setStatus("paid");
    }

    @Override
    public void auth_sms(Order order, String code) {
        System.out.println("Authenticating via SMS");
        this.authenticated = true;
    }

}
