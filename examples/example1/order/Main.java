package example1.order;

public class Main {
    public static void main(String[] args) {
        Order order = new Order();
        order.addItem("Keyboard", 1, 50.0);
        order.addItem("SSD", 1, 150.0);
        order.addItem("USB cable", 2, 5.0);

        System.out.println("Total Price: " + order.totalPrice());

        order.pay("debit", "0372846");
        System.out.println("Order status: " + order.getStatus());
    }
}
