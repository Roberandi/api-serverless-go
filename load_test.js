import http from 'k6/http';
import { check, sleep } from 'k6';

export default function () {
  // Tu URL exacta con la ruta del endpoint
  const url = 'https://x6y3bn8hr5.execute-api.us-east-1.amazonaws.com/notifications/send';
  
  const payload = JSON.stringify({
    email: "test@test.com",
    subject: "Prueba de Carga K6",
    message: "Mensaje de carga desde prueba de estrés"
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(url, payload, params);

  // EL DETECTIVE: Si AWS no nos da un 200, nos dirá exactamente por qué
  if (res.status !== 200) {
    console.log(`\n[ERROR DETECTADO] Status Code: ${res.status} | Respuesta de AWS: ${res.body}\n`);
  }

  check(res, {
    'status fue 200': (r) => r.status === 200,
  });

  sleep(3);
}