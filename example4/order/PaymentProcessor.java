package example4.order;

public abstract class PaymentProcessor {
    public abstract void pay(Order order);

    public abstract void auth_sms(Order order, String code);
}
