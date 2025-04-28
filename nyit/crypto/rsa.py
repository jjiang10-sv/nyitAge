prefix = "0001"
hash = "0640f8d13c0789ff0ed5437cf4bc9f2827d52146dddff38aefc2c17747d45f28"
A = "30 31 30 0D 06 09 60 86 48 01 65 03 04 02 01 05 00 04 20".replace(' ','')
total_len = 256
pad_len = total_len - 1 - (len(A)+len(prefix)+len(hash))//2
prefix + "FF" * pad_len + "00" + A + hash

# from cryptography.hazmat.primitives import hashes
# from cryptography.hazmat.primitives.asymmetric import padding, rsa
# from cryptography.hazmat.primitives import serialization

# # Generate RSA key pair
# private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
# public_key = private_key.public_key()

# # Message to sign
# message = b"I owe you $2000"

# # Sign the message
# signature = private_key.sign(
#     message,
#     padding.PSS(
#         mgf=padding.MGF1(hashes.SHA256()),
#         salt_length=padding.PSS.MAX_LENGTH
#     ),
#     hashes.SHA256()
# )

# # Output the signature
# print("Signature:", signature.hex())