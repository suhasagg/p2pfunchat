from flask import Flask, request
from flask_restful import Resource, Api

app = Flask(__name__)
api = Api(app)



from sentence_transformers import SentenceTransformer, util
model = None


class computeIAB(Resource):
    def get(self,text):
        #Connect to databse
        global model
        import urllib
        text=urllib.parse.unquote(urllib.parse.unquote(text)) 
        sentence = text.split(",")      
        if model is None:
           model = SentenceTransformer('all-MiniLM-L6-v2')

# Two lists of sentences
        sentences1 = [sentence[0]]

        sentences2 = [sentence[1]]

#Compute embedding for both lists
        embeddings1 = model.encode(sentences1, convert_to_tensor=True)
        embeddings2 = model.encode(sentences2, convert_to_tensor=True)

#Compute cosine-similarities
        cosine_scores = util.cos_sim(embeddings1, embeddings2)
        print(cosine_scores)
        return cosine_scores.item()
        


api.add_resource(computeIAB, '/IAB/<string:text>')

if __name__ == '__main__':
     app.run(host='0.0.0.0',port=82,threaded=True)



