package example5.user;

import java.util.HashMap;
import java.util.Map;

class MySQLDatabase {
    private Map<Integer, String> userData;

    public MySQLDatabase() {
        userData = new HashMap<>();
        userData.put(1, "Alice");
        userData.put(2, "Bob");
        userData.put(3, "Charlie");
    }

    public String getUserData(int id) {
        if (userData.containsKey(id)) {
            return "User ID: " + id + ", Name: " + userData.get(id);
        }
        return "User not found";
    }

    public void addUserData(int id, String name) {
        if (userData.containsKey(id)) {
            System.out.println("User ID already exists. Use updateUserData to modify.");
        } else {
            userData.put(id, name);
            System.out.println("User added successfully.");
        }
    }

    public void updateUserData(int id, String name) {
        if (userData.containsKey(id)) {
            userData.put(id, name);
            System.out.println("User updated successfully.");
        } else {
            System.out.println("User ID does not exist. Use addUserData to add new user.");
        }
    }

    public void deleteUserData(int id) {
        if (userData.remove(id) != null) {
            System.out.println("User deleted successfully.");
        } else {
            System.out.println("User ID not found.");
        }
    }
}