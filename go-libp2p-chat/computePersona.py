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
        print(text)
        if model is None:
           #model = SentenceTransformer('all-MiniLM-L6-v2')
           model = SentenceTransformer('all-mpnet-base-v2')
        embeddings = model.encode(text, convert_to_tensor=True)
        category = computeSimilarity(embeddings, model)
        return category

fname = "IABcategories.txt"
with open(fname) as f:
    content = f.readlines()


# get average vector for sentence 2

def computeSimilarity(sentence_1_avg_vector, model):
    i = 0
    highestScore = 0

    similarSentence = ""

    global content

    content1 = [x.strip() for x in content]
    # global modeldata
    # model = modeldata
    for y in content1:
        sentence_2 = y
        i = i + 1
        value = None
        import memcache
        # Cache Category Word vectors in Memcached
        #mc = memcache.Client(['127.0.0.1:11211'], debug=0)

        #    mc.set(y, sentence_2_avg_vector)
        k = y.replace(" ", "").replace(",", "").replace("'", "")
        # print(k)
        # Get category word vectors from memcached
        #value = mc.get(k + "wiki")

        #        print(i)
        # Derive cosine similarity with category word vectors to figure out the category
        if value is not None:
        #Compute cosine-similarities
            cosine_scores = util.cos_sim(sentence_1_avg_vector, value)
        else:
            # if staticVectors.get(y) is None:
            embeddings1 = model.encode(k, convert_to_tensor=True)
            #mc.set(k + "wiki", embeddings1)
            cosine_scores = util.cos_sim(sentence_1_avg_vector, embeddings1)
            #print(cosine_scores)
        if  cosine_scores.float() > highestScore:
            highestScore = cosine_scores.float()
            similarSentence = y
            #print(similarSentence)
    highestScore = 0
    i = 0
    return similarSentence


api.add_resource(computeIAB, '/IABSegments/<string:text>')

if __name__ == '__main__':
     app.run(host='0.0.0.0',port=81,threaded=True)



