public class JavaTest {
    public static void main(String[] args) {
        System.out.println("=== Java Runtime Integration Test ===");
        System.out.println("Java version: " + System.getProperty("java.version"));
        System.out.println("Java home: " + System.getProperty("java.home"));
        System.out.println("Operating system: " + System.getProperty("os.name"));
        System.out.println("SUCCESS: Java runtime is working!");
    }
}