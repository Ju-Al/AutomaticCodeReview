package example5.user;

public class Main {
    public static void main(String[] args) {
        UserService userService = new UserService();
        userService.addUser(1,"Daiana");
        userService.addUser(2, "Tom");
        userService.updateUser(2, "Tom Taylor");
        userService.deleteUser(1);
    }
}
