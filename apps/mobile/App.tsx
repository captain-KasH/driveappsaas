import { StatusBar } from "expo-status-bar";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { config } from "./src/api";
import PassengerApp from "./src/passenger/PassengerApp";
import DriverApp from "./src/driver/DriverApp";

export default function App() {
  return (
    <SafeAreaProvider>
      <StatusBar style="dark" />
      {config.variant === "driver" ? <DriverApp /> : <PassengerApp />}
    </SafeAreaProvider>
  );
}
