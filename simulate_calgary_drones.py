import requests
import time
import random
import math

# Calgary coordinates
CALGARY_LAT = 51.0447
CALGARY_LON = -114.0719

# Backend API endpoint
API_URL = "http://localhost:8080/api/v1/detection"

def generate_drones(num_drones=5):
    """Generate random drone positions around Calgary"""
    drones = []
    for i in range(num_drones):
        # Random offset within ~5km of Calgary center
        lat_offset = random.uniform(-0.05, 0.05)
        lon_offset = random.uniform(-0.05, 0.05)
        
        drone = {
            "SN": f"DRONE_{i+1:03d}",
            "UASID": f"CAL-{i+1:03d}",
            "DroneType": random.choice(["DJI Mavic", "DJI Phantom", "DJI Inspire"]),
            "Latitude": CALGARY_LAT + lat_offset,
            "Longitude": CALGARY_LON + lon_offset,
            "Height": random.uniform(20, 150),
            "Direction": random.randint(0, 359),
            "SpeedHorizontal": random.uniform(0, 15),
            "SpeedVertical": random.uniform(-2, 2),
        }
        drones.append(drone)
    return drones

def send_detection(drone):
    """Send drone detection to backend"""
    try:
        response = requests.post(API_URL, json=drone, timeout=2)
        if response.status_code == 200:
            print(f"✅ Sent {drone['SN']} at ({drone['Latitude']:.4f}, {drone['Longitude']:.4f})")
        else:
            print(f"❌ Failed to send {drone['SN']}: {response.status_code}")
    except Exception as e:
        print(f"❌ Error sending {drone['SN']}: {e}")

def update_drone_position(drone):
    """Update drone position (simulate movement)"""
    # Convert direction to radians
    direction_rad = math.radians(drone["Direction"])
    
    # Move drone based on speed and direction
    # ~111km per degree latitude, ~85km per degree longitude at Calgary's latitude
    speed_kmh = drone["SpeedHorizontal"] * 3.6  # m/s to km/h
    distance_km = (speed_kmh / 3600)  # distance per second in km
    
    drone["Latitude"] += distance_km * math.cos(direction_rad) / 111
    drone["Longitude"] += distance_km * math.sin(direction_rad) / 85
    
    # Random direction change (10% chance)
    if random.random() < 0.1:
        drone["Direction"] = random.randint(0, 359)
    
    # Slight height variation
    drone["Height"] += random.uniform(-2, 2)
    drone["Height"] = max(20, min(150, drone["Height"]))
    
    return drone

def main():
    print("🚁 Calgary Drone Simulator")
    print("=" * 50)
    
    # Generate initial drones
    drones = generate_drones(8)  # 8 drones around Calgary
    
    print(f"Generated {len(drones)} drones around Calgary")
    print("Sending positions every 2 seconds...")
    print("Press Ctrl+C to stop\n")
    
    try:
        while True:
            for drone in drones:
                # Update position
                drone = update_drone_position(drone)
                # Send to backend
                send_detection(drone)
            
            time.sleep(2)  # Update every 2 seconds
            
    except KeyboardInterrupt:
        print("\n\n🛑 Simulation stopped")

if __name__ == "__main__":
    main()