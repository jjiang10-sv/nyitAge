<?php
session_start();
include 'db.php';

if (!isset($_SESSION["user_id"]) || $_SESSION["role"] != "user") {
    header("Location: login.php");
    exit();
}

$user_id = $_SESSION["user_id"];
$upload_success = "";
$update_success = "";

// Extract user_token from the cookie
$cookie_data = isset($_COOKIE["user_token"]) ? explode(":", $_COOKIE["user_token"]) : [];
$cookie_username = $cookie_data[0] ?? null;
$cookie_user_id = $cookie_data[1] ?? null;

// Fetch user details
$query = "SELECT username, email, created_at, role, bank_balance, bio, total_items_purchased, profile_pic, otp FROM users WHERE id = ?";
$stmt = $conn->prepare($query);
$stmt->bind_param("i", $user_id);
$stmt->execute();
$result = $stmt->get_result();
$user = $result->fetch_assoc();
$profile_pic = "uploads/" . ($user["profile_pic"] ?: "default.png");

// Handle profile updates
if ($_SERVER["REQUEST_METHOD"] == "POST" && isset($_POST["update_profile"])) {
    $new_name = trim($_POST["new_name"]);
    $new_bio = trim($_POST["new_bio"]);
    $new_email = trim($_POST["new_email"]);

        $query = "UPDATE users SET username = ?, bio = ?, email = ? WHERE id = ?";
        $stmt = $conn->prepare($query);
        $stmt->bind_param("sssi", $new_name, $new_bio, $new_email, $user_id);
        if ($stmt->execute()) {
            $_SESSION["email"] = $new_email;
            $_SESSION["username"] = $new_name;
            echo "<script>alert('✅ Profile updated successfully!'); window.location.href='user_dashboard.php';</script>";
        } else {
            $update_success = "❌ Failed to update profile.";
        }
    } 


// Handle OTP Reset
if ($_SERVER["REQUEST_METHOD"] == "POST" && isset($_POST["reset_otp"])) {
    $new_otp = rand(100, 999);
    $query = "UPDATE users SET otp = ? WHERE id = ?";
    $stmt = $conn->prepare($query);
    $stmt->bind_param("ii", $new_otp, $user_id);
    if ($stmt->execute()) {
        echo "<script>alert('✅ OTP Reset Successful!'); window.location.href='user_dashboard.php';</script>";
    }
}
?>



<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Dashboard</title>
    <link href="styles.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;600&display=swap');
        body {
            font-family: 'Poppins', sans-serif;
            background: linear-gradient(135deg, #232f3e 0%, #37475a 100%);
            color: #f8f8f8;
            margin: 0;
            padding: 0;
            min-height: 100vh;
        }
        .header-bar {
            width: 100vw;
            height: 8px;
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            box-shadow: 0 2px 8px 0 #ffb34744;
            margin-bottom: 2.5rem;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(15px);
            padding: 40px 30px 30px 30px;
            border-radius: 20px;
            box-shadow: 0px 15px 40px rgba(0, 0, 0, 0.5);
            text-align: center;
            width: 420px;
            max-width: 98vw;
            margin: 60px auto 0 auto;
            animation: fadeInMain 1s cubic-bezier(.77,0,.18,1);
        }
        @keyframes fadeInMain {
            from { opacity: 0; transform: translateY(40px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .profile-pic-container {
            text-align: center;
            margin-bottom: 20px;
        }
        .profile-pic {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            object-fit: cover;
            border: 3px solid #ffe082;
            box-shadow: 0px 0px 15px #ffe08244;
        }
        label {
            display: block;
            text-align: left;
            margin-bottom: 6px;
            margin-top: 18px;
            color: #ffe082;
            font-weight: 600;
        }
        input[type="text"], input[type="email"] {
            width: 95%;
            padding: 12px;
            margin: 8px 0 18px 0;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            background: rgba(255, 255, 255, 0.2);
            color: white;
            transition: background 0.3s, transform 0.2s;
        }
        input[type="text"]:hover, input[type="email"]:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: scale(1.02);
        }
        .otp-btn, .side-btn {
            width: 100%;
            padding: 12px;
            border: none;
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            color: #232f3e;
            font-weight: bold;
            border-radius: 12px;
            cursor: pointer;
            font-size: 16px;
            margin-bottom: 10px;
            transition: background 0.3s, transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 2px 8px 0 #ffb34744;
        }
        .otp-btn:hover, .side-btn:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
            transform: scale(1.05);
            box-shadow: 0 4px 16px 0 #ffb34777;
        }
        .popup {
            display: none;
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: rgba(255, 255, 255, 0.4);
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.3);
        }
        .success {
            color: #ffe082;
            font-size: 1.1rem;
            margin-top: 10px;
        }
        .error {
            color: #ff4f4f;
            font-size: 1.1rem;
            margin-top: 10px;
        }
    </style>
    <script>
        function toggleEdit(field) {
            document.getElementById(field + "_display").style.display = "none";
            document.getElementById(field + "_input").style.display = "block";
            document.getElementById(field + "_edit").style.display = "none";
            document.getElementById(field + "_check").style.display = "inline-block";
        }

        function saveChanges() {
            document.getElementById("profileForm").submit();
        }

        function fetchOTP() {
    fetch('user_dashboard.php', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: 'fetch_otp=1'
    })
    .then(response => response.text())
    .then(data => {
        document.getElementById("otp-display").innerText = "Your OTP: " + data;
    });
}

    </script>
</head>
<body>
    <div class="header-bar"></div>
    <div class="container">
        <h2>Welcome, <?php echo htmlspecialchars($user['username']); ?>! 🎉</h2>

        <div class="profile-pic-container">
            <img src="<?php echo $profile_pic; ?>" alt="Profile Picture" class="profile-pic">
            <button onclick="window.location.href='change_pfp.php'" class="change-pfp-btn">Change Profile Picture</button>
        </div>
        <form method="post">
            <button class="otp-btn" type="submit" name="reset_otp">Reset OTP</button>
        </form>
        <h3 id="otp-display">Your OTP: <?php echo htmlspecialchars($user['otp']); ?></h3>
        <button class="otp-btn" onclick="fetchOTP()">Fetch OTP</button>


        <form method="post" id="profileForm">
            <label>Name: 
                <span id="new_name_display"><?php echo htmlspecialchars($user['username']); ?></span>
                <button type="button" id="new_name_edit" class="edit-btn" onclick="toggleEdit('new_name')">✏️</button>
                <button type="button" id="new_name_check" class="check-btn" onclick="saveChanges()" style="display: none;">✔️</button>
            </label>
            <input type="text" id="new_name_input" name="new_name" value="<?php echo htmlspecialchars($user['username']); ?>" style="display: none;">

            <label>Bio: 
                <span id="new_bio_display"><?php echo htmlspecialchars($user['bio']); ?></span>
                <button type="button" id="new_bio_edit" class="edit-btn" onclick="toggleEdit('new_bio')">✏️</button>
                <button type="button" id="new_bio_check" class="check-btn" onclick="saveChanges()" style="display: none;">✔️</button>
            </label>
            <textarea id="new_bio_input" name="new_bio" style="display: none;"><?php echo htmlspecialchars($user['bio']); ?></textarea>

            <label>Email: 
    <span id="email_display"><?php echo htmlspecialchars($user['email']); ?></span>
    <button type="button" class="change-email-btn" onclick="openEmailPopup()">✉️ Change Email</button>
</label>

            <input type="submit" id="saveBtn" name="update_profile" value="Save Changes">
        </form>

<!-- Popup for Changing Email -->
<div id="emailPopup" class="popup">
    <div class="popup-content">
        <h2>Change Email</h2>
        <p><strong>Current Email:</strong> <span id="current-email"><?php echo htmlspecialchars($user['email']); ?></span></p>
        <form action="change_email.php" method="post">
            <label>New Email:</label>
            <input type="email" name="new_email" required>
            <input type="hidden" name="username" value="<?php echo htmlspecialchars($user['username']); ?>">
            <button type="submit">Submit</button>
            <button type="button" onclick="closeEmailPopup()">Cancel</button>
        </form>
    </div>
</div>

<script>
    function openEmailPopup() {
        document.getElementById("emailPopup").style.display = "block";
    }

    function closeEmailPopup() {
        document.getElementById("emailPopup").style.display = "none";
    }
</script>

        <h3>Account Details</h3>
        <p><strong>Account Created:</strong> <?php echo htmlspecialchars($user['created_at']); ?></p>
        <p><strong>Role:</strong> <?php echo htmlspecialchars(ucfirst($user['role'])); ?></p>
        <p><strong>Bank Balance:</strong> $<?php echo number_format($user['bank_balance'], 2); ?></p>
        <p><strong>Total Items Purchased:</strong> <?php echo $user['total_items_purchased']; ?></p>

        <?php if (!empty($update_success)) { echo "<p class='success'>$update_success</p>"; } ?>
        <button onclick="window.location.href='index.php'" class="side-btn">🏠 Go to Homepage</button>

        <button onclick="window.location.href='logout.php'" class="logout-btn">Logout</button>
    </div>
</body>
</html>
