import { Button } from '@/components/ui/button';
import { getSensorData, getSensors } from '@/lib/api';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';

export function MainPage() {
  const [selectedSensor, setSelectedSensor] = useState<number | null>(null);

  const { data: sensors } = useQuery({
    queryKey: ['sensor-data'],
    queryFn: getSensors
  });

  const { data: sensorData } = useQuery({
    queryKey: ['sensor-data', selectedSensor],
    // biome-ignore lint/style/noNonNullAssertion: <explanation>
    queryFn: () => getSensorData(selectedSensor!),
    enabled: !!selectedSensor,
  });

  return (
    <div className="flex h-dvh">
      <div className="border-2 w-48 flex justify-center py-16">
        {sensors?.map((sensor: number) => (
          <Button type="button" key={sensor} onClick={() => setSelectedSensor(sensor)}>
            Sensor {sensor}
          </Button>
        ))}
      </div>

      <div className="flex items-center justify-center flex-grow p-4">
        {sensorData ? (
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={sensorData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="timestamp"
                tickFormatter={(tick) => new Date(tick * 1000).toLocaleTimeString()} // Convert Unix timestamp to human-readable time
              />
              <YAxis />
              <Tooltip labelFormatter={(label) => new Date(label * 1000).toLocaleTimeString()} />
              <Legend />
              <Line type="monotone" dataKey="temperature" stroke="#ff7300" />
              <Line type="monotone" dataKey="humidity" stroke="#387908" />
              <Line type="monotone" dataKey="luminosity" stroke="#8884d8" />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <p>Select a sensor to view data</p>
        )}
      </div>
    </div>
  );
}
