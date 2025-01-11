package example4.order;

public class Main {
    public static void main(String[] args) {
        Order order = new Order();

        order.addItem("Keyboard", 1, 50.0);
        order.addItem("SSD", 1, 150.0);
        order.addItem("USB cable", 2, 5.0);

        PaymentProcessor processor = new CreditCardPaymentProcessor("0372846");
        PaymentProcessor paypalProcessor = new DebitCardPaymentProcessor("123456");
        processor.pay(order);
        System.out.println("Order status: " + order.getStatus());
        paypalProcessor.auth_sms(order, "123456");
        System.out.println("Order status: " + order.getStatus());
    }
}
