const API_URL = import.meta.env.VITE_API_URL;

export const getSensors = async () => {
	const response = await fetch(`http://${API_URL}/sensors`);

	console.log(API_URL);

	return response.json();
};

export const getSensorData = async (sensorId: number) => {
	const response = await fetch(
		`http://${API_URL}/sensors/${sensorId}/readings`,
	);

	return response.json();
};
