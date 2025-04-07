import { useNavigate } from "react-router-dom";

const LandingPage = () => {
  const navigate = useNavigate();

  return (
    <div className="container text-center mt-5">
      <h1>Welcome to NISHAD's Social Media App</h1>
      <div className="mt-4">
        <button className="btn btn-primary m-2" onClick={() => navigate("/login")}>
          Login
        </button>
        <button className="btn btn-success m-2" onClick={() => navigate("/signup")}>
          Sign Up
        </button>
      </div>
    </div>
  );
};

export default LandingPage;
