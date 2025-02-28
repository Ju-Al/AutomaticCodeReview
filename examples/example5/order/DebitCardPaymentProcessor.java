package example5.order;

public class DebitCardPaymentProcessor extends PaymentProcessor {
    private SMSAuthorizer authorizer;
    private String securityCode;

    public DebitCardPaymentProcessor(String securityCode, SMSAuthorizer authorizer) {
        this.authorizer = authorizer;
        this.securityCode = securityCode;
    }

    @Override
    public void pay(Order order) {
        if (!this.authorizer.is_authentificated()) {
            throw new IllegalStateException("Not authentificated");
        }
        System.out.println("Processing debit card payment");
        System.out.println("Verifying security code: " + this.securityCode);
        order.setStatus("paid");
    }

}
