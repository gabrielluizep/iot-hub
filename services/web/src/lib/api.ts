const API_URL = import.meta.env.VITE_API_URL;

export const getSensors = async () => {
	const response = await fetch(`${API_URL}/sensors`);

	return response.json();
};

export const getSensorData = async (sensorId: number) => {
	const response = await fetch(`${API_URL}/sensors/${sensorId}/readings`);

	return response.json();
};

export const postSensorLight = async (sensorId: number, lightOn: boolean) => {
	const response = await fetch(`${API_URL}/sensors/${sensorId}`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ lightOn }),
	});

	return response.json();
};