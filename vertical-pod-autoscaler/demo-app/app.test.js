const request = require('supertest');

const app = require('./app');

describe('GET /', () => {
  it('GET / => array of stuffs', () => {
    return request(app)
      .get('/stuff')

      .expect('Content-Type', /json/)

      .expect(200)

      .then((response) => {
        expect(response.body).toEqual(
            expect.arrayContaining([
                expect.objectContaining({
                    _id: expect.any(String),
                    title: expect.any(String),
                    description: expect.any(String),
                    imageUrl: expect.any(String),
                    price: expect.any(Number),
                    userId: expect.any(String)

            }),
          ])
        );
      });
  });

  it('GET /id => 404 if stuff not found', () => {
    return request(app).get('/10000000000').expect(404);
  });

});