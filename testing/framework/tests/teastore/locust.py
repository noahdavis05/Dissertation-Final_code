import random
from locust import HttpUser, task, between, tag

class TeaStoreUser(HttpUser):
    # User waits between 1 and 5 seconds per request
    wait_time = between(1, 5)

    def on_start(self):
        self.client.get("/tools.descartes.teastore.webui/")

    @tag('front-end')
    @task(10)
    def view_home(self):
        self.client.get("/tools.descartes.teastore.webui/", name="Home")

    @tag('cpu-intensive')
    @task(5)
    def login_stress(self):
        """stresses auth service - should use CPU"""
        self.client.post(
            "/tools.descartes.teastore.webui/loginAction",
            data={"username": "user1", "password": "password"},
            name="Login"
        )

    @tag('memory-intensive')
    @task(8)
    def browse_products(self):
        """stresses image and recomender services"""
        category_id = random.randint(2, 6)
        self.client.get(
            f"/tools.descartes.teastore.webui/category?id={category_id}", 
            name="Category"
        )
        product_id = random.randint(1, 100)
        self.client.get(
            f"/tools.descartes.teastore.webui/product?id={product_id}", 
            name="ProductDetail"
        )

    @tag('io-intensive')
    @task(3)
    def view_cart(self):
        """stresses recommender and persistence"""
        self.client.get("/tools.descartes.teastore.webui/cart", name="Cart")