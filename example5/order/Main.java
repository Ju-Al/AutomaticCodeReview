package example5.order;

public class Main {
    public static void main(String[] args) {
        Order order = new Order();

        order.addItem("Keyboard", 1, 50.0);
        order.addItem("SSD", 1, 150.0);
        order.addItem("USB cable", 2, 5.0);

        SMSAuthorizer authorizer = new SMSAuthorizer();
        authorizer.verify_code("12345");

        DebitCardPaymentProcessor processor = new DebitCardPaymentProcessor("0372846", authorizer);
        processor.pay(order);

        System.out.println(order.getStatus());
    }
}
