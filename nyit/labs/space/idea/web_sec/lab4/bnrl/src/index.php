<?php
session_start();
include 'db.php';

// Fetch products from the database
$query = "SELECT id, name, description, price, image, rating FROM products";
$result = $conn->query($query);
$products = $result->fetch_all(MYSQLI_ASSOC);

// Check if user is logged in
$is_logged_in = isset($_SESSION["user_id"]);
$user_pfp = $is_logged_in ? "uploads/" . ($_SESSION["profile_pic"] ?? "default.png") : "";
$user_id = $is_logged_in ? $_SESSION["user_id"] : null;
?>

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Buy Now Regret Later - Home</title>
    
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
            margin-bottom: 1.5rem;
        }
        .navbar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: #111;
            padding: 18px 32px;
            color: white;
            border-radius: 0 0 1.5rem 1.5rem;
            box-shadow: 0 4px 24px 0 #232f3e55;
        }
        .navbar a {
            color: white;
            text-decoration: none;
            margin: 0 18px;
            font-size: 19px;
            font-weight: bold;
            transition: color 0.3s;
            letter-spacing: 0.5px;
        }
        .navbar a:hover {
            color: #ffb347;
        }
        .navbar-right {
            display: flex;
            align-items: center;
        }
        .profile-pic {
            width: 44px;
            height: 44px;
            border-radius: 50%;
            margin-left: 18px;
            border: 2.5px solid #ffb347;
            box-shadow: 0 2px 8px 0 #ffb34744;
        }
        .container {
            max-width: 1800px;
            margin: auto;
            padding: 20px 0 40px 0;
        }
        .main-content {
            display: flex;
            flex-wrap: wrap;
            gap: 2.5rem;
            justify-content: center;
            align-items: flex-start;
            padding: 2rem 2vw 2rem 2vw;
            width: 100%;
            min-height: 60vh;
            box-sizing: border-box;
            animation: fadeInMain 1s cubic-bezier(.77,0,.18,1);
        }
        @keyframes fadeInMain {
            from { opacity: 0; transform: translateY(40px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .product-card {
            background: linear-gradient(135deg, #2d3a4e 60%, #3e5c76 100%);
            border-radius: 2rem;
            box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37), 0 1.5px 8px 0 #ffb34744;
            padding: 2.5rem 2rem 2rem 2rem;
            margin: 0.5rem 0;
            width: 350px;
            max-width: 95vw;
            display: flex;
            flex-direction: column;
            align-items: center;
            transition: transform 0.35s cubic-bezier(.77,0,.18,1), box-shadow 0.35s cubic-bezier(.77,0,.18,1);
            animation: cardPopIn 0.7s cubic-bezier(.77,0,.18,1);
            animation-fill-mode: backwards;
            position: relative;
        }
        .product-card:not(:first-child) {
            animation-delay: 0.15s;
        }
        @keyframes cardPopIn {
            from { opacity: 0; transform: scale(0.95) translateY(40px); }
            to { opacity: 1; transform: scale(1) translateY(0); }
        }
        .product-card:hover {
            transform: translateY(-10px) scale(1.03) rotateZ(-0.5deg);
            box-shadow: 0 16px 48px 0 #ffb34755, 0 2px 12px 0 #232f3e99;
            z-index: 2;
        }
        .product-card img {
            width: 180px;
            height: 180px;
            object-fit: cover;
            border-radius: 1.5rem;
            margin-bottom: 1.5rem;
            box-shadow: 0 4px 24px 0 #ffb34733;
            transition: transform 0.3s cubic-bezier(.77,0,.18,1);
        }
        .product-card:hover img {
            transform: scale(1.07) rotateZ(2deg);
        }
        .product-card h3 {
            font-size: 1.5rem;
            font-weight: 600;
            margin: 0.5rem 0 0.7rem 0;
            color: #ffe082;
            letter-spacing: 0.5px;
        }
        .product-card p {
            font-size: 1.08rem;
            color: #e0e0e0;
            margin: 0.5rem 0 1.2rem 0;
            min-height: 48px;
        }
        .product-card strong {
            font-size: 1.3rem;
            font-weight: 700;
            color: #ffb347;
            margin-bottom: 0.5rem;
        }
        .product-card .rating {
            font-size: 1.1rem;
            color: #ffe082;
            margin-bottom: 1.2rem;
        }
        .add-to-cart, .rev {
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            color: #232f3e;
            font-weight: 700;
            border: none;
            border-radius: 1rem;
            padding: 0.8rem 1.5rem;
            font-size: 1.08rem;
            cursor: pointer;
            box-shadow: 0 2px 8px 0 #ffb34744;
            transition: background 0.25s, transform 0.18s, box-shadow 0.25s;
            outline: none;
            margin: 0.2rem 0.5rem 0.2rem 0.5rem;
            letter-spacing: 0.5px;
        }
        .add-to-cart:hover, .rev:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
            transform: scale(1.08) translateY(-2px);
            box-shadow: 0 4px 16px 0 #ffb34777;
        }
        .add-to-cart:active, .rev:active {
            transform: scale(0.97);
            box-shadow: 0 1px 4px 0 #ffb34755;
        }
        .review-section {
            display: none;
            background: rgba(255, 255, 255, 0.1);
            padding: 15px;
            border-radius: 12px;
            text-align: left;
            margin-top: 15px;
            border: 1px solid rgba(255, 255, 255, 0.2);
            box-shadow: 0px 4px 12px rgba(0, 0, 0, 0.2);
            max-width: 500px;
            margin-left: auto;
            margin-right: auto;
        }
        .review-section h4 {
            font-size: 16px;
            margin-bottom: 10px;
            color: #ffb347;
        }
        .review {
            display: flex;
            align-items: center;
            justify-content: space-between;
            background: rgba(0, 0, 0, 0.2);
            padding: 10px;
            border-radius: 8px;
            margin-bottom: 8px;
        }
        .r-btn, .edit-btn, .delete-btn {
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            color: #232f3e;
            border: none;
            padding: 8px 12px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            font-weight: bold;
            transition: background 0.3s, transform 0.2s;
            display: flex;
            align-items: center;
            gap: 5px;
            margin-left: 6px;
        }
        .r-btn:hover, .edit-btn:hover, .delete-btn:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
            transform: scale(1.08) translateY(-2px);
        }
        .delete-btn {
            color: #ff4f4f;
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
        }
        .delete-btn:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
        }
        textarea {
            width: 100%;
            padding: 8px;
            border-radius: 8px;
            background: rgba(255, 255, 255, 0.2);
            color: white;
            border: none;
            outline: none;
            font-size: 15px;
            resize: none;
            margin-top: 10px;
            margin-bottom: 8px;
            transition: background 0.3s, box-shadow 0.3s;
        }
        textarea:focus {
            background: rgba(255, 255, 255, 0.3);
            box-shadow: 0 0 0 2px #ffb34755;
        }
        textarea::placeholder {
            color: rgba(255, 255, 255, 0.6);
        }
        @media (max-width: 900px) {
            .main-content {
                flex-direction: column;
                align-items: center;
                gap: 2rem;
                padding: 2rem 1vw 1rem 1vw;
            }
            .product-card {
                width: 95vw;
                max-width: 420px;
            }
            .navbar {
                flex-direction: column;
                padding: 18px 8px;
            }
        }
        @media (min-width: 901px) {
            .product-card {
                animation: floatCard 4s ease-in-out infinite alternate;
            }
            @keyframes floatCard {
                from { box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37), 0 1.5px 8px 0 #ffb34744; }
                to { box-shadow: 0 16px 48px 0 #ffb34755, 0 2px 12px 0 #232f3e99; }
            }
        }
    </style>
</head>
<body>
    <div class="header-bar"></div>
    <div class="navbar">
        <div class="navbar-left">
            <a href="index.php">🏠 Home</a>
            <a href="cart.php">🛒 Cart</a>
            <?php if ($is_logged_in) { ?>
                <a href="user_dashboard.php">👤 Dashboard</a>
            <?php } ?>
        </div>
        <div class="navbar-right">
            <?php if ($is_logged_in) { ?>
                <img src="<?php echo $user_pfp; ?>" alt="Profile Picture" class="profile-pic">
                <a href="logout.php">🚪 Logout</a>
            <?php } else { ?>
                <a href="login.php">🔑 Login</a>
                <a href="register.php">✍ Register</a>
            <?php } ?>
        </div>
    </div>

    <div class="container">
        <h1 style="text-align:center; color:#ffe082; font-size:2.3rem; margin-bottom:0.5em; letter-spacing:1px;">Welcome to Buy Now Regret Later 🛒</h1>
        <p style="text-align:center; color:#e0e0e0; font-size:1.15rem; margin-bottom:2.5em;">The only store where you buy now and question your choices later! Featuring the dumbest, most useless, and absolutely regrettable products!</p>
        <div class="main-content">
            <?php foreach ($products as $product) { ?>
                <div class="product-card">
                    <img src="uploads/<?php echo $product['image']; ?>" alt="<?php echo $product['name']; ?>">
                    <h3><?php echo $product['name']; ?></h3>
                    <p><?php echo $product['description']; ?></p>
                    <p><strong>$<?php echo number_format($product['price'], 2); ?></strong></p>
                    <p>⭐ <?php echo number_format($product['rating'], 1); ?> / 5</p>
                    <button class="add-to-cart" onclick="addToCart(<?php echo $product['id']; ?>)">Add to Cart</button>
                    <button class="rev" onclick="toggleReviews(<?php echo $product['id']; ?>)">Check Reviews</button>

                    <div id="reviews-<?php echo $product['id']; ?>" class="review-section">
                        <h4>Reviews:</h4>

                        <?php 
                        // Fetch reviews for this product
                        $review_query = "SELECT r.id, r.user_id, r.review_text, u.username 
                                         FROM reviews r 
                                         JOIN users u ON r.user_id = u.id 
                                         WHERE r.product_id = ?";
                        $stmt = $conn->prepare($review_query);
                        $stmt->bind_param("i", $product['id']);
                        $stmt->execute();
                        $review_result = $stmt->get_result();
                        $reviews = $review_result->fetch_all(MYSQLI_ASSOC);
                        ?>

                        <div id="review-list-<?php echo $product['id']; ?>">
                            <?php if (!empty($reviews)) { ?>
                                <?php foreach ($reviews as $review) { ?>
                                    <div class='review' id="review-<?php echo $review['id']; ?>">
                                        <span id="review-text-<?php echo $review['id']; ?>">
                                            <?php echo htmlspecialchars($review['review_text']); ?>
                                        </span> 
                                        <small>- <?php echo htmlspecialchars($review['username']); ?></small>

                                        <?php if ($is_logged_in && $user_id == $review['user_id']) { ?>
                                            <button class="edit-btn" onclick="editReview(<?php echo $review['id']; ?>, <?php echo $product['id']; ?>)">✏️</button>
                                            <button class="delete-btn" onclick="deleteReview(<?php echo $review['id']; ?>, <?php echo $product['id']; ?>)">🗑️</button>
                                        <?php } ?>
                                    </div>
                                <?php } ?>
                            <?php } else { ?>
                                <p>No reviews yet. Be the first to regret your purchase!</p>
                            <?php } ?>
                        </div>

                        <?php if ($is_logged_in) { ?>
                            <textarea id="review-text-<?php echo $product['id']; ?>" placeholder="Write a review..."></textarea>
                            <button class="rev" onclick="submitReview(<?php echo $product['id']; ?>)">Submit Review</button>
                        <?php } ?>
                    </div>
                </div>
            <?php } ?>
        </div>
    </div>

    <script>
        let gg = "https://pastebin.com/92mKVNBB";
        const LOGGED_IN_USER_ID = <?php echo isset($_SESSION["user_id"]) ? $_SESSION["user_id"] : 'null'; ?>;
        console.log("Logged-in User ID:", LOGGED_IN_USER_ID);

        function addToCart(productId) {
            fetch('add_to_cart.php', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: 'product_id=' + productId
            })
            .then(response => response.text())
            .then(data => alert(data));
        }
        
        function toggleReviews(productId) {
            const reviewSection = document.getElementById("reviews-" + productId);
            if (reviewSection.style.display === "none" || reviewSection.style.display === "") {
                reviewSection.style.display = "block";
                fetchReviews(productId);
            } else {
                reviewSection.style.display = "none";
            }
        }

        function fetchReviews(productId) {
            fetch('fetch_reviews.php?product_id=' + productId)
            .then(response => response.json())
            .then(data => {
                let reviewList = document.getElementById("review-list-" + productId);
                reviewList.innerHTML = "";

                data.forEach(review => {
                    let reviewHTML = `
                        <div class="review" id="review-${review.id}">
                            <span id="review-text-${review.id}">${review.review_text}</span> 
                            <small>- ${review.username}</small>
                    `;

                    // Verify owner in JavaScript
                    console.log("Review ID:", review.id, "User ID:", review.user_id, "Logged-in ID:", LOGGED_IN_USER_ID);

                    if (LOGGED_IN_USER_ID !== null && LOGGED_IN_USER_ID == review.user_id) { 
                        reviewHTML += `
                            <button class="r-btn" onclick="editReview(${review.id}, ${productId})">✏️ Edit</button>
                            <button class="r-btn" onclick="deleteReview(${review.id}, ${productId})">🗑️ Delete</button>
                        `;
                    }
                    else{
                        console.log("failed to get the ID");
                    }

                    reviewHTML += `</div>`;
                    reviewList.innerHTML += reviewHTML;
                });
            });
        }

        function editReview(reviewId, productId) {
            let reviewText = document.getElementById("review-text-" + reviewId).innerHTML;
            let newText = prompt("Edit your review:", reviewText);

            if (newText !== null && newText.trim() !== "") {
                fetch('edit_review.php', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: `review_id=${reviewId}&review_text=${encodeURIComponent(newText)}`
                })
                .then(response => response.text())
                .then(data => {
                    alert(data);
                    if (data.includes("✅")) {
                        document.getElementById("review-text-" + reviewId).innerHTML = newText;
                    }
                });
            }
        }

        function deleteReview(reviewId, productId) {
            if (confirm("Are you sure you want to delete this review?")) {
                fetch('delete_review.php', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: `review_id=${reviewId}`
                })
                .then(response => response.text())
                .then(data => {
                    alert(data);
                    if (data.includes("✅")) {
                        document.getElementById("review-" + reviewId).remove();
                    }
                });
            }
        }

        function submitReview(productId) {
            const reviewText = document.getElementById("review-text-" + productId).value;
            fetch('submit_review.php', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: `product_id=${productId}&review_text=${encodeURIComponent(reviewText)}`
            })
            .then(response => response.text())
            .then(data => {
                alert(data);
                fetchReviews(productId);
            });
        }
    </script>
</body>
</html>