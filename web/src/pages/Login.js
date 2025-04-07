import { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";

const Login = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");

  const handleLogin = async (e) => {
    e.preventDefault();
    setMessage("");
    
    try {
      const response = await axios.post("http://localhost:8080/v1/authentication/token", {
        email,
        password,
      });
      console.log(response) 
      localStorage.setItem("jwtToken", response.data.data);
      console.log("response: ", response)
      console.log("response data: ", response.data) 
      console.log("Saved Token:", localStorage.getItem("jwtToken"));
      // setMessage("Login Successful!");
      navigate("/home");
    } catch (err) {
      setMessage("Invalid email or password");
    }
  };

  return (
    <div className="container mt-5">
      <div className="row justify-content-center">
        <div className="col-md-4">
          <h2 className="text-center">Login</h2>
          {message && <div className={`alert ${message === "Login Successful!" ? "alert-success" : "alert-danger"}`}>{message}</div>}
          <form onSubmit={handleLogin}>
            <div className="mb-3">
              <label className="form-label">Email</label>
              <input
                type="email"
                className="form-control"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="mb-3">
              <label className="form-label">Password</label>
              <input
                type="password"
                className="form-control"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <button type="submit" className="btn btn-primary w-100">Login</button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default Login;
