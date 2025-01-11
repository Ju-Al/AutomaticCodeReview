package example3.order;

public interface PaymentProcessor {
    public void pay(Order order, String securityCode);
}
