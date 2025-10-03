import React, { useState, useEffect } from "react";

// --- Interface สำหรับข้อมูลผู้ใช้ ---
interface User {
  status: string;
  status_desc: string;
  username: string;
  firstname: string;
  lastname: string;
  costcenter: string;
  ba: string;
  part: string;
  area: string;
  job_name: string;
  level: string;
  div_name: string;
  dep_name: string;
  org_name: string;
  email: string;
  position: string;
}

// --- สไตล์ Tailwind CSS ในรูปแบบ Object เพื่อความกระชับ ---
const styles: { [key: string]: string } = {
  container:
    "min-h-screen bg-gray-100 flex flex-col items-center justify-center font-sans p-4",
  card: "bg-white shadow-2xl rounded-xl p-8 md:p-12 w-full max-w-md transition-all duration-300",
  title: "text-3xl font-bold text-center text-gray-800 mb-2",
  subtitle: "text-center text-gray-500 mb-8",
  form: "space-y-6",
  inputGroup: "relative",
  input:
    "w-full px-4 py-3 bg-gray-50 border-2 border-gray-200 rounded-lg text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all",
  button:
    "w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-4 rounded-lg transition-transform transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-gray-400 disabled:cursor-not-allowed",
  errorMessage: "mt-4 text-center text-red-600 bg-red-100 p-3 rounded-lg",
  loggedInContainer: "text-center",
  welcomeMessage: "text-2xl font-semibold text-gray-800",
  userInfo: "text-gray-600 mt-2",
  logoutButton:
    "mt-8 bg-red-600 hover:bg-red-700 text-white font-bold py-2 px-6 rounded-lg transition-transform transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500",
};

// --- Main App Component ---
function App() {
  // State สำหรับเก็บข้อมูลในฟอร์ม
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  // State สำหรับเก็บข้อมูลผู้ใช้ที่ login แล้ว
  const [user, setUser] = useState<User | null>(null);

  // State สำหรับจัดการ Loading และ Error
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  // Hook นี้จะทำงานตอน Component โหลดขึ้นมาครั้งแรก
  // เพื่อเช็คว่ามีข้อมูล user ใน Local Storage หรือไม่
  useEffect(() => {
    try {
      const loggedInUserJSON = localStorage.getItem("user");
      if (loggedInUserJSON) {
        const userData: User = JSON.parse(loggedInUserJSON);
        setUser(userData);
      }
    } catch (e) {
      console.error("Failed to parse user data from localStorage", e);
      localStorage.removeItem("user");
    }
  }, []);

  // --- ฟังก์ชันจัดการการ Login ---
  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); // ป้องกันการ refresh หน้า
    setLoading(true);
    setError("");

    try {
      const response = await fetch(
        "https://intranet.pwa.co.th/login/webservice_login6.php",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            username: username,
            pwd: password,
          }),
        },
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data: User = await response.json();

      if (data.status === "success") {
        localStorage.setItem("user", JSON.stringify(data));
        setUser(data);
      } else {
        setError(data.status_desc || "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง");
      }
    } catch (err) {
      console.error("Login API error:", err);
      setError("เกิดข้อผิดพลาดในการเชื่อมต่อ กรุณาลองใหม่อีกครั้ง");
    } finally {
      setLoading(false);
    }
  };

  // --- ฟังก์ชันจัดการการ Logout ---
  const handleLogout = () => {
    localStorage.removeItem("user");
    setUser(null);
  };

  // --- ส่วนของการแสดงผล (Render) ---

  if (user) {
    return (
      <div className={styles.container}>
        <div className={styles.card}>
          <div className={styles.loggedInContainer}>
            <p className={styles.subtitle}>เข้าสู่ระบบสำเร็จ</p>
            <h1 className={styles.welcomeMessage}>
              ยินดีต้อนรับ, {user.firstname} {user.lastname}
            </h1>
            <p className={styles.userInfo}>ตำแหน่ง: {user.position}</p>
            <p className={styles.userInfo}>อีเมล: {user.email}</p>
            <button onClick={handleLogout} className={styles.logoutButton}>
              ออกจากระบบ
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h1 className={styles.title}>PWA Intranet Login</h1>
        <p className={styles.subtitle}>กรุณาลงชื่อเข้าเพื่อใช้งานระบบ</p>

        <form onSubmit={handleLogin} className={styles.form}>
          <div className={styles.inputGroup}>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                setUsername(e.target.value)
              }
              placeholder="Username"
              className={styles.input}
              required
            />
          </div>
          <div className={styles.inputGroup}>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                setPassword(e.target.value)
              }
              placeholder="Password"
              className={styles.input}
              required
            />
          </div>

          {error && <div className={styles.errorMessage}>{error}</div>}

          <div>
            <button type="submit" className={styles.button} disabled={loading}>
              {loading ? "กำลังตรวจสอบ..." : "เข้าสู่ระบบ"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default App;
